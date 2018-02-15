package main

import (
	"fmt"
	"time"

	"./actions"
	"./signal"
	v "./variable"
)

func main() {
	actions.Debug = false

	cpu := v.Variable{Name: "cpus"}
	ram := v.Variable{Name: "memory", Unit: "m"}
	io := v.Variable{Name: "blkio-weight"}

	var sine1 signal.Signal = signal.SineFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    85 * time.Second,
			Amplitude: 0.95,
			Offset:    0.05,
			DutyCycle: 1,
			Variable:  &cpu,
		},
	}

	var sine2 signal.Signal = signal.SineFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    110 * time.Second,
			Amplitude: 60,
			Offset:    4,
			DutyCycle: 1,
			Variable:  &ram,
		},
	}

	var sine3 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    135 * time.Second,
			Amplitude: 990,
			Offset:    10,
			DutyCycle: 1,
			Variable:  &io,
		},
	}

	fmt.Println("Starting...")

	var stressContainer actions.IAction = &actions.LimitContainer{}
	stressContainer.Default()
	stressContainer.AddVariable(&cpu)
	var stressContainer2 actions.IAction = &actions.LimitContainer{}
	stressContainer2.Default()
	stressContainer2.AddVariable(&ram)
	var stressContainer3 actions.IAction = &actions.LimitContainer{AsInt: true}
	stressContainer3.Default()
	stressContainer3.AddVariable(&io)

	var manager = signal.NewManager("main")
	manager.AddMonkey(signal.NewChaosMonkey("sampler1", append([]*signal.Signal{}, &sine1), 0*time.Second, 5*time.Second, stressContainer))
	manager.AddMonkey(signal.NewChaosMonkey("sampler2", append([]*signal.Signal{}, &sine2), 25*time.Second, 5*time.Second, stressContainer2))
	manager.AddMonkey(signal.NewChaosMonkey("sampler3", append([]*signal.Signal{}, &sine3), 50*time.Second, 5*time.Second, stressContainer3))

	manager.GenerateScripts()

	fmt.Println("Done!")
}
