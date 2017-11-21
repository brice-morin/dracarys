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
  actions.Action
  Ticker* time.Ticker
}

/*type Sampler interface {
  NewSampler(Signal, int) SamplerParams
  Start(c chan int) //will emit on channel c the sampled value (from Signal) every Rate seconds
  Stop()
}*/

func NewChaosMonkey(sig Signal, r time.Duration, a actions.Action) ChaosMonkey {
  var sampler ChaosMonkey = ChaosMonkey{
    Signal: sig,
    Rate: r,
    Action: a,
  }
  return sampler
}

func (s *ChaosMonkey) Start(/*c chan Sample*/) {
  var now int64 = time.Now().Unix()
  //var ts int64 = 0
  //c <- Sample{ts, s.Signal.Sample(ts),}
  //var ticker = time.NewTicker(s.Rate)

  s.Ticker = time.NewTicker(s.Rate)
  //go func() {
    for t := range s.Ticker.C {
        ts := t.Unix() - now
        v := s.Signal.Sample(ts)
        s.Action.Do(v)
        /*c <- Sample{
          ts,
          s.Signal.Sample(ts),
        }*/
    }
  //}()
}

func (s *ChaosMonkey) Stop() {
  s.Ticker.Stop()
}
