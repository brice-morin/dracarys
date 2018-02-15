package signal

type Square struct {
	PeriodicSignal
}

func (s Square) Sample(t int64) float64 {
	if t%int64(s.Period.Seconds()) < int64(s.Period.Seconds()*s.DutyCycle) {
		return s.Amplitude + s.Offset
	}
	return s.Offset
}
