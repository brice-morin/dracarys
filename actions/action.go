package actions

import "time"

type IAction interface {
	Do(int64)
	GetTick() time.Duration
	SetTick(time.Duration)
}

type Action struct {
	IAction
	Duration time.Duration
}

func (a *Action) SetTick(t time.Duration) {
	a.Duration = t
}

func (a *Action) GetTick() time.Duration {
	return a.Duration
}
