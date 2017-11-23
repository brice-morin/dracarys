package main

import (
	"fmt"
	"time"

	"./docker"
	"./signal"
)

func main() {
	var sine signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second, //FIXME: this should be a time.Duration
			Amplitude:  100,
			Offset:     0,
			Duty_cycle: 1.0,
		},
	}

	fmt.Println("Starting...")
	//var restart docker.RestartContainer
	var packetLoss docker.PacketLoss
	//packetLoss.Do(100)
	//var sampler = signal.NewChaosMonkey(sine, 120*time.Second, restart)
	var sampler2 = signal.NewChaosMonkey(sine, 30*time.Second, packetLoss)

	//sampler.Start()
	sampler2.Start()

}
