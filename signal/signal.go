package signal

import (
	"time"

	v "../variable"
)

type Signal interface {
	/**
	 * Returns the sampled value between -Amplitude and +Amplitude (or 0 and +Amplitude) for this signal at a given time.
	 * This time should be <= Period
	 */
	Sample(int64) float64
	GetPeriod() time.Duration
	GetVariable() *v.Variable
}

type PeriodicSignal struct {
	Period    time.Duration
	Amplitude float64
	Offset    float64
	DutyCycle float64
	Variable  *v.Variable
}

func (p PeriodicSignal) GetPeriod() time.Duration {
	return p.Period
}

func (p PeriodicSignal) GetVariable() *v.Variable {
	return p.Variable
}
