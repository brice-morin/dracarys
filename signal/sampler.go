package signal

import (
	"container/ring"
	"time"

	"../actions"
)

type Sample struct {
	t int64
	v int64
}

type ChaosMonkey struct {
	samples *ring.Ring
	Signal
	Rate time.Duration
	actions.IAction
	Ticker *time.Ticker
}

func NewChaosMonkey(sig *Signal, r time.Duration, a actions.IAction) *ChaosMonkey {
	a.SetTick(r)
	var sampler ChaosMonkey = ChaosMonkey{
		Signal:  *sig,
		Rate:    r,
		IAction: a,
	}
	return &sampler
}

func (s *ChaosMonkey) Init() ChaosMonkey {
	size := int(s.Signal.GetPeriod().Seconds()) / int(s.Rate.Seconds())
	s.samples = ring.New(size)
	for i := 0; i < int(s.Signal.GetPeriod().Seconds()); i = i + int(s.Rate.Seconds()) {
		s.samples.Value = s.Signal.Sample(int64(i))
		s.samples = s.samples.Next()
	}
	s.Signal = nil
	return *s
}

func (s *ChaosMonkey) Start(c chan Sample, done *chan bool) {
	var now int64 = time.Now().Unix()
	s.Ticker = time.NewTicker(s.Rate)
	for t := range s.Ticker.C {
		ts := t.Unix() - now
		v := s.samples.Value.(int64)
		s.samples = s.samples.Next()
		//v := s.Signal.Sample(ts)
		s.IAction.Do(v)
		c <- Sample{
			ts,
			v,
			//s.Signal.Sample(ts),
		}
	}
	*done <- true
}

func (s *ChaosMonkey) Stop() {
	s.Ticker.Stop()
}
