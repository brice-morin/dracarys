package main

import (
	"fmt"
	"time"

	"./actions"
	"./signal"
)

func main() {
	actions.Debug = false

	/*var sine1 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  10000,
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
	}*/

	var sine4 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     600 * time.Second,
			Amplitude:  1000,
			Offset:     1, //for scaling services up and down, offset should be initial number of instances so as to move around this initial value
			Duty_cycle: 1,
		},
	}

	var sine5 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  30,
			Offset:     0, //for scaling services up and down, offset should be initial number of instances so as to move around this initial value
			Duty_cycle: 0.6,
		},
	}

	/*var sine6 signal.Signal = signal.Sine{
		PeriodicSignal: signal.PeriodicSignal{
			Period:     180 * time.Second,
			Amplitude:  20,
			Offset:     0, //for scaling services up and down, offset should be initial number of instances so as to move around this initial value
			Duty_cycle: 0.75,
		},
	}*/

	fmt.Println("Starting...")
	//done := make(chan bool)

	/*var networkRate docker.NetworkRate
	var packetLoss docker.PacketLoss
	var packetDelay docker.PacketDelay
	var restartContainer docker.RestartContainer*/
	var stressContainer actions.IAction = &actions.StressContainer{Resource: "cpu"}
	/*var stressContainer2 actions.IAction = &actions.StressContainer{Resource: "vm"}
	var stressContainer3 actions.IAction = &actions.StressContainer{Resource: "io"}
	var stressContainer4 actions.IAction = &actions.StressContainer{Resource: "hdd"}*/

	var networkRate actions.IAction = &actions.NetworkRate{}

	var sampler = signal.NewChaosMonkey(&sine5, 30*time.Second, stressContainer)
	//c := make(chan signal.Sample)
	/*var sampler2 = signal.NewChaosMonkey(&sine5, 30*time.Second, stressContainer2).Init()
	c2 := make(chan signal.Sample)
	var sampler3 = signal.NewChaosMonkey(&sine5, 30*time.Second, stressContainer3).Init()
	c3 := make(chan signal.Sample)
	var sampler4 = signal.NewChaosMonkey(&sine5, 30*time.Second, stressContainer4).Init()
	c4 := make(chan signal.Sample)*/

	var sampler5 = signal.NewChaosMonkey(&sine4, 10*time.Second, networkRate)
	//c5 := make(chan signal.Sample)

	sampler.GenerateScript("sampler.sh")
	sampler5.GenerateScript("sampler5.sh")
	/*go sampler.Start(c, &done)
	go sampler2.Start(c2, &done)
	go sampler3.Start(c3, &done)
	go sampler4.Start(c4, &done)

	go sampler5.Start(c5, &done)

	go func() {
		for {
			select {
			case s := <-c:
				fmt.Println(time.Now().Unix(), ": Stress container", s)
			case s2 := <-c2:
				fmt.Println(time.Now().Unix(), ": Stress container2", s2)
			case s3 := <-c3:
				fmt.Println(time.Now().Unix(), ": Stress container3", s3)
			case s4 := <-c4:
				fmt.Println(time.Now().Unix(), ": Stress container4", s4)
			case s5 := <-c5:
				fmt.Println(time.Now().Unix(), ": Network rate", s5)
			}
		}
	}()

	<-done //Wait for chaos monkeys to terminate (never)*/

}
