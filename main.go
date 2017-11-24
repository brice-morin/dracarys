package main

import (
	"fmt"
	"time"

	"./docker"
	"./signal"
)

func main() {
	docker.Debug = true

	var sine1 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  1000,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine2 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  100,
			Offset:     0,
			Duty_cycle: 0.5,
		},
	}

	var sine3 signal.Signal = signal.Sine{ //Sine is absolute value of sine, while sine2 is the real sine (with + and - values)
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  500,
			Offset:     0,
			Duty_cycle: 0.75,
		},
	}

	var sine4 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     600 * time.Second,
			Amplitude:  4,
			Offset:     1, //for scaling services up and down, offset should be initial number of instances so as to move around this initial value
			Duty_cycle: 1,
		},
	}

	fmt.Println("Starting...")
	done := make(chan bool)

	var networkRate docker.NetworkRate
	var packetLoss docker.PacketLoss
	var packetDelay docker.PacketDelay
	var restartContainer docker.RestartContainer

	var sampler1 = signal.NewChaosMonkey(&sine1, 30*time.Second, &networkRate)
	c1 := make(chan signal.Sample)
	var sampler2 = signal.NewChaosMonkey(&sine2, 90*time.Second, &packetLoss)
	c2 := make(chan signal.Sample)
	var sampler3 = signal.NewChaosMonkey(&sine3, 50*time.Second, &packetDelay)
	c3 := make(chan signal.Sample)
	var sampler4 = signal.NewChaosMonkey(&sine4, 180*time.Second, &restartContainer)
	c4 := make(chan signal.Sample)

	go sampler1.Start(c1, &done)
	go sampler2.Start(c2, &done)
	go sampler3.Start(c3, &done)
	go sampler4.Start(c4, &done)

	go func() {
		for {
			select {
			case s1 := <-c1:
				fmt.Println(time.Now().Unix(), ": Network Rate", s1)
			case s2 := <-c2:
				fmt.Println(time.Now().Unix(), ": Packet Loss", s2)
			case s3 := <-c3:
				fmt.Println(time.Now().Unix(), ": Packet delay", s3)
			case s4 := <-c4:
				fmt.Println(time.Now().Unix(), ": Container restart", s4)
			}
		}
	}()

	<-done //Wait for chaos monkeys to terminate (never)

}
