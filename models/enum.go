package models

import "database/sql/driver"

type RunningState int

const (
	Running RunningState = 1
	Paused  RunningState = -1
)

func (p *RunningState) Scan(value interface{}) error {
	*p = RunningState(value.(int))
	return nil
}

func (p RunningState) Value() (driver.Value, error) {
	return int(p), nil
}
