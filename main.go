package main

import (
  "fmt"
  "time"
  "./signal"
  "./docker"
)

func main() {
  var sine signal.Signal = signal.Sine{
    PeriodicSignal: signal.PeriodicSignal {
      Period: 60*time.Second, //FIXME: this should be a time.Duration
      Amplitude: 3,
      Offset: 0,
      Duty_cycle: 0.5,
    },
  }

  fmt.Println("Starting...")
  var restart docker.RestartContainer
  var sampler = signal.NewChaosMonkey(sine, 2*time.Second, restart)
  /*c := make(chan signal.Sample)

  go func () {
    for sample := range c {
      fmt.Println(sample)
    }
  }()*/

  sampler.Start(/*c*/)

  /*time.Sleep(time.Second * 5)
  sampler.Stop()
  fmt.Println("Stopping")*/

  /*var kill docker.KillContainer
  kill.Do(1)*/



}
