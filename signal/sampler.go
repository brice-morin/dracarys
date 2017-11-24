package signal

import (
	"time"

	"../actions"
)

type Sample struct {
	t int64
	v int64
}

type ChaosMonkey struct {
	Signal
	Rate time.Duration
	actions.IAction
	Ticker *time.Ticker
}

func NewChaosMonkey(sig *Signal, r time.Duration, a actions.IAction) ChaosMonkey {
	a.SetTick(r)
	var sampler ChaosMonkey = ChaosMonkey{
		Signal:  *sig,
		Rate:    r,
		IAction: a,
	}
	return sampler
}

func (s *ChaosMonkey) Start(c chan Sample, done *chan bool) {
	var now int64 = time.Now().Unix()
	s.Ticker = time.NewTicker(s.Rate)
	for t := range s.Ticker.C {
		ts := t.Unix() - now
		v := s.Signal.Sample(ts)
		s.IAction.Do(v)
		c <- Sample{
			ts,
			s.Signal.Sample(ts),
		}
	}
	*done <- true
}

func (s *ChaosMonkey) Stop() {
	s.Ticker.Stop()
}
