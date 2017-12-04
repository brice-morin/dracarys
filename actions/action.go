package actions

import (
	"time"
)

type IAction interface {
	Do(int64)                       // Do something to all containers
	_do(int64, func(int64, string)) //Do something (apply func) to all containers. FIXME: this is a hack to simulate "proper" dispatch with inheritance...
	DoTo(int64, string)             //Do something to a specific ContainerRestart
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
