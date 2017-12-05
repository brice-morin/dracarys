package actions

import (
	"time"
)

type IAction interface {
	GetTick() time.Duration
	SetTick(time.Duration)
	Print() string
}

type Action struct {
	Duration time.Duration
}

func (a *Action) SetTick(t time.Duration) {
	a.Duration = t
}

func (a *Action) GetTick() time.Duration {
	return a.Duration
}
