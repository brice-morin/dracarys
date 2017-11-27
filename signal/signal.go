package signal

import (
	"time"
)

type Signal interface {
	/**
	 * Returns the sampled value between -Amplitude and +Amplitude (or 0 and +Amplitude) for this signal at a given time.
	 * This time should be <= Period
	 */
	Sample(int64) int64
	GetPeriod() time.Duration
}

type PeriodicSignal struct {
	Period     time.Duration
	Amplitude  int64
	Offset     int64
	Duty_cycle float64
}

func (p PeriodicSignal) GetPeriod() time.Duration {
	return p.Period
}
