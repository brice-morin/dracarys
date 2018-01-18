package actions

import (
	"time"
)

type Target uint8
type Scope uint8

const (
	CONTAINER Target = iota
	SERVICE
	NETWORK
)

const (
	ALL Scope = iota
	RND
	SOME
)

type IAction interface {
	Default() IAction
	GetTick() time.Duration
	SetTick(time.Duration)
	GetScope() Scope
	SetScope(Scope)
	GetTarget() Target
	SetTarget(Target)
	Print() string
	PrintHelper() string
	GetResources() []string
	AddResource(string)
}

type Action struct {
	Target    Target
	Scope     Scope
	Duration  time.Duration
	Resources []string
}

func (a *Action) PrintHelper() string {
	return ""
}

func (a *Action) SetTick(t time.Duration) {
	a.Duration = t
}

func (a *Action) GetTick() time.Duration {
	return a.Duration
}

func (a *Action) SetTarget(t Target) {
	a.Target = t
}

func (a *Action) GetTarget() Target {
	return a.Target
}

func (a *Action) SetScope(s Scope) {
	a.Scope = s
}

func (a *Action) GetScope() Scope {
	return a.Scope
}

func (a *Action) AddResource(r string) {
	if a.Scope == SOME {
		a.Resources = append(a.Resources, r)
	}
}

func (a *Action) GetResources() []string {
	if a.Scope == SOME {
		return a.Resources
	}
	return nil
}
