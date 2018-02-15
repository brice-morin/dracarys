package signal

import (
	"math"
)

type Bessel struct {
	PeriodicSignal
}

//Returns 0 if bessel(x) < 0
type Bessel2 struct {
	PeriodicSignal
}

//Returns |bessel(x)| if bessel(x) < 0
type Bessel3 struct {
	PeriodicSignal
}

func (s Bessel) Sample(t int64) int64 {
	return int64(s.Amplitude*math.J0(float64(t))) + int64(s.Offset)
}

func (s Bessel2) Sample(t int64) int64 {
	var bessel = int64(s.Amplitude*math.J0(float64(t))) + int64(s.Offset)
	if bessel < 0 {
		return 0
	}
	return bessel
}

func (s Bessel3) Sample(t int64) int64 {
	var bessel = int64(s.Amplitude*math.J0(float64(t))) + int64(s.Offset)
	if bessel < 0 {
		return int64(math.Abs(float64(bessel)))
	}
	return bessel
}
