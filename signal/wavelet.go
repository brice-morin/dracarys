package signal

import (
	"math"
)

type Wavelet struct {
	PeriodicSignal
}

type WaveletFloat struct {
	PeriodicSignal
}

func sampleWavelet(p float64, a float64, o float64, t int64) float64 {
	var pt float64 = float64(t % int64(p))
	if pt != 0 && int64(pt)%int64(p/2) == 0 {
		return float64(a + o)
	}
	var x float64 = pt - p/2
	var c float64 = 0.1
	return (a / 2 * (math.Sin(float64(c*2*x)) - math.Sin(float64(c*x))) / float64(c*x)) + a/2 + o
}

func (s WaveletFloat) Sample(t int64) float64 {
	return sampleWavelet(s.Period.Seconds(), s.Amplitude, s.Offset, t)
}

func (s Wavelet) Sample(t int64) int64 {
	return int64(sampleWavelet(s.Period.Seconds(), s.Amplitude, s.Offset, t))
}
