package models

import "database/sql/driver"

type RunningState int64

const (
	RSRunning RunningState = 1
	RSPaused  RunningState = -1
)

func (p *RunningState) Scan(value interface{}) error {
	*p = RunningState(value.(int64))
	return nil
}

func (p RunningState) Value() (driver.Value, error) {
	return int64(p), nil
}
