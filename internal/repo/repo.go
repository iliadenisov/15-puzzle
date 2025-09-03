package repo

import (
	"15-puzzle/internal/model"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path"
	"slices"
	"sync"
	"time"
)

type FileRepo struct {
	dataFile string
	latch    sync.RWMutex
	data     *model.Data
}

func NewFileRepo(ctx context.Context, dataFile string) (*FileRepo, error) {
	dataFile = path.Clean(dataFile)
	r := &FileRepo{
		dataFile: dataFile,
		data:     &model.Data{Users: make(map[int]model.User)},
	}

	b, err := os.ReadFile(dataFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if len(b) == 0 {
		// init with empty writeable file
		if err := r.withData(func(d *model.Data) {}); err != nil {
			return nil, err
		}
		return r, nil
	}

	if err := json.Unmarshal(b, &r.data); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *FileRepo) Monitoring() (model.Monitoring, error) {
	r.latch.RLock()
	defer r.latch.RUnlock()

	m := &model.Monitoring{Users: len(r.data.Users)}
	for i := range r.data.Users {
		m.GamesStarted += r.data.Users[i].GamesStarted
		m.GamesSolved += r.data.Users[i].GamesSolved
	}
	return *m, nil
}

func (r *FileRepo) Stats(UserID int) (model.User, error) {
	r.latch.RLock()
	defer r.latch.RUnlock()

	user, ok := r.data.Users[UserID]
	if !ok {
		return model.User{}, fmt.Errorf("user not found: user_id=%d", UserID)
	}
	return user, nil
}

func (r *FileRepo) AddUser(UserID int) error {
	if err := r.withUser(UserID, func(*model.User) {}); err != nil {
		return err
	}
	return nil
}

func (r *FileRepo) RegisterGameStart(UserID int) (model.User, error) {
	var result model.User
	if err := r.withUser(UserID, func(u *model.User) {
		u.GamesStarted++
		ts := model.JSONTimestamp(time.Now().UTC())
		u.LastStartTime = &ts
		result = *u
	}); err != nil {
		return result, err
	}
	return result, nil
}

func (r *FileRepo) RegisterGameSolve(UserID, moves int) (model.User, error) {
	var result model.User
	if err := r.withUser(UserID, func(u *model.User) {
		u.GamesSolved++
		if u.LastStartTime != nil {
			moveAverage := float32(time.Since(time.Time(*u.LastStartTime)).Seconds() / float64(moves))
			if u.BestResult == nil || moveAverage < *u.BestResult {
				ts := model.JSONTimestamp(time.Now().UTC())
				u.BestSolveTime = &ts
				u.BestResult = &moveAverage
				u.LastStartTime = nil
			}
		}
		result = *u
	}); err != nil {
		return result, err
	}
	return result, nil
}

func (r *FileRepo) Rating() []int {
	r.latch.RLock()
	defer r.latch.RUnlock()

	return slices.SortedFunc(maps.Keys(r.data.Users), func(a, b int) int { return SortRating(r.data.Users[a], r.data.Users[b]) })
}

func SortRating(a, b model.User) int {
	return cmp.Or(
		compareBest(a.BestResult, b.BestResult),     // lower value is higher
		compareTs(a.BestSolveTime, b.BestSolveTime), // earlier value is higher
		cmp.Compare(b.GamesSolved, a.GamesSolved),   // greater value is higher
		cmp.Compare(a.UserID, b.UserID),             // sadly, compare user ids when nobody has solved just once
	)
}

func compareBest(a, b *float32) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return +1
	case b == nil:
		return -1
	default:
		return cmp.Compare(*a, *b)
	}
}

func compareTs(a, b *model.JSONTimestamp) int {
	switch {
	case a == nil && b == nil:
		return 0
	case a == nil:
		return +1
	case b == nil:
		return -1
	default:
		return time.Time(*a).Compare(time.Time(*b))
	}
}

func (r *FileRepo) withUser(userID int, acceptor func(d *model.User)) error {
	return r.withData(func(d *model.Data) {
		user, ok := d.Users[userID]
		if !ok {
			user = model.User{UserID: userID}
		}
		acceptor(&user)
		d.Users[userID] = user
	})
}

func (r *FileRepo) withData(acceptor func(d *model.Data)) error {
	r.latch.Lock()
	defer r.latch.Unlock()

	acceptor(r.data)

	b, err := json.Marshal(r.data)
	if err != nil {
		return fmt.Errorf("data marshall: %s", err)
	}
	repoDir := path.Dir(r.dataFile)
	tf, err := os.CreateTemp(repoDir, "15-puzzle")
	if err != nil {
		return fmt.Errorf("create temp file at %s: %s", repoDir, err)
	}

	if _, err := tf.Write(b); err != nil {
		return fmt.Errorf("write %s: %s", tf.Name(), err)
	}

	if err := tf.Close(); err != nil {
		return fmt.Errorf("close %s to %s: %s", tf.Name(), r.dataFile, err)
	}

	if err := os.Rename(tf.Name(), r.dataFile); err != nil {
		return fmt.Errorf("rename %s to %s: %s", tf.Name(), r.dataFile, err)
	}

	return nil
}
