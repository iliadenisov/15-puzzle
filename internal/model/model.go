package model

import (
	"encoding/json"
	"time"
)

type JSONTimestamp time.Time

func (t JSONTimestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).Unix())
}

func (t *JSONTimestamp) UnmarshalJSON(data []byte) error {
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*t = JSONTimestamp(time.Unix(v, 0))
	return nil
}

type Data struct {
	Users map[int]User `json:"users"`
}

type User struct {
	UserID        int            `json:"user_id"`
	GamesStarted  int            `json:"games_started"`
	GamesSolved   int            `json:"games_solved"`
	LastStartTime *JSONTimestamp `json:"last_start_ts,omitempty"`
	BestResult    *float32       `json:"best_result,omitempty"`
	BestSolveTime *JSONTimestamp `json:"best_solve_ts,omitempty"`
	Monitoring    *Monitoring    `json:"-"`
}

type ApiResponse struct {
	Stats      *Stats      `json:"stats,omitempty"`
	Monitoring *Monitoring `json:"monitoring,omitempty"`
	Info       *Info       `json:"info,omitempty"`
	Err        *string     `json:"error,omitempty"`
}

type Stats struct {
	Rank         int `json:"rank"`
	GamesStarted int `json:"games_started"`
	GamesSolved  int `json:"games_solved"`
}

type Monitoring struct {
	Users        int `json:"users,omitempty"`
	GamesStarted int `json:"games_started,omitempty"`
	GamesSolved  int `json:"games_solved,omitempty"`
}

type Info struct {
	ProjectLink string `json:"project_link"`
}
