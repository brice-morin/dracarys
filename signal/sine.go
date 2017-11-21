package signal

import (
  "math"
)

type Sine struct {
  PeriodicSignal
}

//amplitude*|sin(PI*t/(period*duty_cycle))|
func (s Sine) Sample(t int64) int64 {
  var pt int64 = t % int64(s.Period.Seconds())
  if (pt < int64(s.Period.Seconds() * s.Duty_cycle)) {
    return int64(float64(s.Amplitude) * math.Abs(math.Sin(math.Pi*(float64(pt) / float64(float64(s.Period.Seconds()) * s.Duty_cycle)))) + float64(s.Offset))
  } else {
    return 0
  }
}
