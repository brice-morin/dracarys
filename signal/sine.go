package signal

import (
	"math"
)

type Sine struct {
	PeriodicSignal
}

type SineFloat struct {
	PeriodicSignal
}

type Sine2 struct {
	PeriodicSignal
}

//amplitude*|sin(PI*t/(period*duty_cycle))|
func (s SineFloat) Sample(t int64) float64 {
	var pt int64 = t % int64(s.Period.Seconds())
	if pt < int64(s.Period.Seconds()*s.DutyCycle) {
		return float64(s.Amplitude)*math.Abs(math.Sin(math.Pi*(float64(pt)/float64(float64(s.Period.Seconds())*s.DutyCycle)))) + float64(s.Offset)
	}
	return 0
}

//amplitude*|sin(PI*t/(period*duty_cycle))|
func (s Sine) Sample(t int64) float64 {
	var pt int64 = t % int64(s.Period.Seconds())
	if pt < int64(s.Period.Seconds()*s.DutyCycle) {
		return float64(int64(float64(s.Amplitude)*math.Abs(math.Sin(math.Pi*(float64(pt)/float64(float64(s.Period.Seconds())*s.DutyCycle)))) + float64(s.Offset)))
	}
	return 0
}

//amplitude*sin(PI*t/(period*duty_cycle))
func (s Sine2) Sample(t int64) int64 {
	var pt int64 = t % int64(s.Period.Seconds())
	if pt < int64(s.Period.Seconds()*s.DutyCycle) {
		return int64(float64(s.Amplitude)*math.Sin(math.Pi*(float64(pt)/float64(float64(s.Period.Seconds())*s.DutyCycle))) + float64(s.Offset))
	}
	return 0
}
