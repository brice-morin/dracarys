package actions

import (
	"time"

	v "../variable"
)

type Type uint8
type Scope uint8

const (
	CONTAINER Type = iota
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
	GetType() Type
	SetType(Type)
	Print() string
	PrintHelper() string
	GetTargets() []string
	AddTarget(string)
	GetVariables() []*v.Variable
	AddVariable(*v.Variable)
}

type Action struct {
	Type      Type
	Scope     Scope
	Duration  time.Duration
	Variables []*v.Variable
	Targets   []string
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

func (a *Action) SetType(t Type) {
	a.Type = t
}

func (a *Action) GetType() Type {
	return a.Type
}

func (a *Action) SetScope(s Scope) {
	a.Scope = s
}

func (a *Action) GetScope() Scope {
	return a.Scope
}

func (a *Action) AddTarget(r string) {
	if a.Scope == SOME {
		a.Targets = append(a.Targets, r)
	} else {
		panic(a)
	}
}

func (a *Action) GetTargets() []string {
	return a.Targets
}

func (a *Action) AddVariable(v *v.Variable) {
	a.Variables = append(a.Variables, v)
}

func (a *Action) GetVariables() []*v.Variable {
	return a.Variables
}
