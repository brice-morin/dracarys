package main

import (
	"fmt"
	"time"

	"./docker"
	"./signal"
)

func main() {
	var sine1 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second, //FIXME: this should be a time.Duration
			Amplitude:  1000,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine2 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second, //FIXME: this should be a time.Duration
			Amplitude:  100,
			Offset:     0,
			Duty_cycle: 0.5,
		},
	}

	var sine3 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second, //FIXME: this should be a time.Duration
			Amplitude:  500,
			Offset:     0,
			Duty_cycle: 0.75,
		},
	}

	fmt.Println("Starting...")
	done := make(chan bool)

	var networkRate docker.NetworkRate
	var packetLoss docker.PacketLoss
	var packetDelay docker.PacketDelay

	var sampler1 = signal.NewChaosMonkey(&sine1, 10*time.Second, &networkRate)
	c1 := make(chan signal.Sample)
	var sampler2 = signal.NewChaosMonkey(&sine2, 30*time.Second, &packetLoss)
	c2 := make(chan signal.Sample)
	var sampler3 = signal.NewChaosMonkey(&sine3, 15*time.Second, &packetDelay)
	c3 := make(chan signal.Sample)

	go sampler1.Start(c1, &done)
	go sampler2.Start(c2, &done)
	go sampler3.Start(c3, &done)

	go func() {
		for s1 := range c1 {
			fmt.Println(time.Now().Unix(), ": Network Rate", s1)
		}
	}()

	go func() {
		for s2 := range c2 {
			fmt.Println(time.Now().Unix(), ": Packet Loss", s2)
		}
	}()

	go func() {
		for s3 := range c3 {
			fmt.Println(time.Now().Unix(), ": Packet delay", s3)
		}
	}()

	<-done //Wait for chaos monkeys to terminate (never)

}
