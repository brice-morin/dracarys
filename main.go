package main

import (
	"fmt"
	"time"

	"./actions"
	"./signal"
)

func main() {
	actions.Debug = false

	var sine1 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:     600 * time.Second,
			Amplitude:  30,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine2 signal.Signal = signal.Sine{ //Kill, scale, disconnect, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:     300 * time.Second,
			Amplitude:  15,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine3 signal.Signal = signal.Sine{ //Network rate
		PeriodicSignal: signal.PeriodicSignal{
			Period:     180 * time.Second,
			Amplitude:  500,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine4 signal.Signal = signal.Sine{ //Network delay
		PeriodicSignal: signal.PeriodicSignal{
			Period:     180 * time.Second,
			Amplitude:  1000,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	var sine5 signal.Signal = signal.Sine{ //Network loss
		PeriodicSignal: signal.PeriodicSignal{
			Period:     180 * time.Second,
			Amplitude:  100,
			Offset:     0,
			Duty_cycle: 1,
		},
	}

	fmt.Println("Starting...")

	var stressContainer actions.IAction = &actions.StressContainer{Resource: "cpu"}
	stressContainer.Default()
	var stressContainer2 actions.IAction = &actions.StressContainer{Resource: "vm"}
	stressContainer2.Default()
	var stressContainer3 actions.IAction = &actions.StressContainer{Resource: "hdd"}
	stressContainer3.Default()
	var stressContainer4 actions.IAction = &actions.StressContainer{Resource: "io"}
	stressContainer4.Default()

	var killContainer actions.IAction = &actions.KillContainer{}
	killContainer.Default()
	var scaleService actions.IAction = &actions.ScaleService{}
	scaleService.Default()
	scaleService.SetScope(actions.SOME)
	scaleService.AddResource("diversiot_device_0-0-0")
	var disconnectNetwork actions.IAction = &actions.NetworkDisconnect{}
	disconnectNetwork.Default()

	var networkRate actions.IAction = &actions.NetworkRate{}
	networkRate.Default()
	var networkDelay actions.IAction = &actions.PacketDelay{}
	networkDelay.Default()
	var networkLoss actions.IAction = &actions.PacketLoss{}
	networkLoss.Default()

	var manager = signal.NewManager("main")
	manager.AddMonkey(signal.NewChaosMonkey("sampler", &sine1, 0*time.Second, 30*time.Second, stressContainer))
	manager.AddMonkey(signal.NewChaosMonkey("sampler2", &sine1, 150*time.Second, 30*time.Second, stressContainer2))
	manager.AddMonkey(signal.NewChaosMonkey("sampler3", &sine1, 300*time.Second, 30*time.Second, stressContainer3))
	manager.AddMonkey(signal.NewChaosMonkey("sampler4", &sine1, 450*time.Second, 30*time.Second, stressContainer4))

	manager.AddMonkey(signal.NewChaosMonkey("sampler5", &sine2, 0*time.Second, 20*time.Second, killContainer))
	manager.AddMonkey(signal.NewChaosMonkey("sampler6", &sine2, 100*time.Second, 20*time.Second, scaleService))
	manager.AddMonkey(signal.NewChaosMonkey("sampler7", &sine2, 200*time.Second, 20*time.Second, disconnectNetwork))

	manager.AddMonkey(signal.NewChaosMonkey("sampler8", &sine3, 0*time.Second, 10*time.Second, networkRate))
	manager.AddMonkey(signal.NewChaosMonkey("sampler9", &sine4, 60*time.Second, 10*time.Second, networkDelay))
	manager.AddMonkey(signal.NewChaosMonkey("sampler10", &sine5, 120*time.Second, 10*time.Second, networkLoss))

	manager.GenerateScripts()

	fmt.Println("Done!")
}
