package repo_test

import (
	"15-puzzle/internal/model"
	"15-puzzle/internal/repo"
	"context"
	"maps"
	"math/rand/v2"
	"os"
	"slices"
	"testing"
	"time"
)

func ref(t time.Time) *model.JSONTimestamp {
	if t.IsZero() {
		return nil
	}
	r := model.JSONTimestamp(t)
	return &r
}

func TestSortRating(t *testing.T) {
	fp := func(v float32) *float32 { return &v }
	users := map[int]model.User{
		1: {UserID: 1, BestResult: fp(5), GamesStarted: 7, GamesSolved: 5, BestSolveTime: ref(time.Now().Add(-time.Second * 3))},
		2: {UserID: 2, BestResult: fp(5), GamesStarted: 7, GamesSolved: 4, BestSolveTime: ref(time.Now().Add(-time.Second * 3))},
		3: {UserID: 3, BestResult: fp(5), GamesStarted: 7, GamesSolved: 3, BestSolveTime: ref(time.Now().Add(-time.Second * 3))},
		4: {UserID: 4, BestResult: fp(5), GamesStarted: 4, GamesSolved: 3, BestSolveTime: ref(time.Now().Add(-time.Second * 2))},
		5: {UserID: 5, BestResult: fp(5), GamesStarted: 4, GamesSolved: 3, BestSolveTime: ref(time.Now().Add(-time.Second * 1))},
		6: {UserID: 6, BestResult: fp(6), GamesStarted: 4, GamesSolved: 3, BestSolveTime: ref(time.Now().Add(-time.Second * 1))},
		7: {UserID: 7, BestResult: fp(7), GamesStarted: 3, GamesSolved: 0, BestSolveTime: ref(time.Time{})},
	}
	keys := slices.Collect(maps.Keys(users))
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })
	result := slices.SortedFunc(slices.Values(keys), func(a, b int) int { return repo.SortRating(users[a], users[b]) })
	expected := []int{1, 2, 3, 4, 5, 6, 7}
	if slices.Compare(expected, result) != 0 {
		t.Errorf("rating not correct\nexpected: %v\nactual: %v", expected, result)
	}
}

func TestRepo(t *testing.T) {
	testWithNewRepo(t, func(t *testing.T, r *repo.FileRepo) { assertRating(t, []int{}, r) })
	testWithNewRepo(t, testPlayersAndRatings)
}

func testWithNewRepo(t *testing.T, test func(t *testing.T, r *repo.FileRepo)) {
	f, err := os.CreateTemp("", "puzzle15-repo-test")
	if err != nil {
		t.Fatalf("temporary file create: %s", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("temporary file close: %s", err)
	}
	if err := os.Remove(f.Name()); err != nil {
		t.Fatalf("temporary file remove before: %s", err)
	}
	defer func() {
		if err := os.Remove(f.Name()); err != nil {
			t.Errorf("temporary file remove after: %s", err)
		}
	}()

	testWithRepo(t, f.Name(), test)
}

func testWithRepo(t *testing.T, file string, test func(t *testing.T, r *repo.FileRepo)) {
	r, err := repo.NewFileRepo(context.Background(), file)
	if err != nil {
		t.Fatalf("repo init: %s", err)
	}

	test(t, r)
}

func testPlayersAndRatings(t *testing.T, r *repo.FileRepo) {
	assertRating(t, []int{}, r)

	testUserOne := 1
	assertRegisterGameStart(t, testUserOne, r,
		model.User{UserID: testUserOne, GamesStarted: 1, GamesSolved: 0, BestSolveTime: ref(time.Time{})})
	assertRating(t, []int{testUserOne}, r)

	testUserTwo := 2
	assertRegisterGameStart(t, testUserTwo, r,
		model.User{UserID: testUserTwo, GamesStarted: 1, GamesSolved: 0, BestSolveTime: ref(time.Time{})})
	assertRating(t, []int{testUserOne, testUserTwo}, r)

	testUserThree := 3
	assertRegisterGameStart(t, testUserThree, r,
		model.User{UserID: testUserThree, GamesStarted: 1, GamesSolved: 0, BestSolveTime: ref(time.Time{})})
	assertRegisterGameSolve(t, testUserThree, 30, r,
		model.User{UserID: testUserThree, GamesStarted: 1, GamesSolved: 1, BestSolveTime: ref(time.Now())})
	assertRating(t, []int{testUserThree, testUserOne, testUserTwo}, r)

	assertRegisterGameSolve(t, testUserTwo, 20, r,
		model.User{UserID: testUserTwo, GamesStarted: 1, GamesSolved: 1, BestSolveTime: ref(time.Now())})
	assertRating(t, []int{testUserThree, testUserTwo, testUserOne}, r)

	assertRegisterGameStart(t, testUserTwo, r,
		model.User{UserID: testUserTwo, GamesStarted: 2, GamesSolved: 1, BestSolveTime: ref(time.Now())})
	assertRegisterGameSolve(t, testUserTwo, 10, r,
		model.User{UserID: testUserTwo, GamesStarted: 2, GamesSolved: 2, BestSolveTime: ref(time.Now())})
	assertRating(t, []int{testUserThree, testUserTwo, testUserOne}, r)
}

func assertUserHaveValues(t *testing.T, expected, actual model.User) {
	if expected.UserID != actual.UserID {
		t.Errorf("expect UserID=%d, actual: %d", expected.UserID, actual.UserID)
	}
	if expected.GamesStarted != actual.GamesStarted {
		t.Errorf("expect GamesStarted=%d, actual: %d", expected.GamesStarted, actual.GamesStarted)
	}
	if expected.GamesSolved != actual.GamesSolved {
		t.Errorf("expect GamesSolved=%d, actual: %d", expected.GamesSolved, actual.GamesSolved)
	}
	expectedBst := time.Time{}
	if expected.BestSolveTime != nil {
		expectedBst = time.Time(*expected.BestSolveTime)
	}
	actualLst := time.Time{}
	if actual.BestSolveTime != nil {
		actualLst = time.Time(*actual.BestSolveTime)
	}
	var delta time.Duration = time.Millisecond * 20
	if actualLst.After(expectedBst.Add(delta*time.Millisecond)) ||
		actualLst.Before(expectedBst.Add(-delta*time.Millisecond)) {
		t.Errorf("expect BestSolveTime=%v, actual: %v, delta: %d", expectedBst, actualLst, delta)
	}
}

func assertRegisterGameStart(t *testing.T, UserID int, r *repo.FileRepo, expected model.User) {
	u, err := r.RegisterGameStart(UserID)
	if err != nil {
		t.Fatalf("RegisterGameStart: %s", err)
	}
	assertUserHaveValues(t, expected, u)
}

func assertRegisterGameSolve(t *testing.T, UserID, moves int, r *repo.FileRepo, expected model.User) {
	u, err := r.RegisterGameSolve(UserID, moves)
	if err != nil {
		t.Fatalf("RegisterGameSolve: %s", err)
	}
	assertUserHaveValues(t, expected, u)
}

func assertRating(t *testing.T, expected []int, r *repo.FileRepo) {
	actual := r.Rating()
	if slices.Compare(expected, actual) != 0 {
		t.Errorf("get rating: wrong value\nexpected: %#v\nactual: %#v", expected, actual)
	}
}
