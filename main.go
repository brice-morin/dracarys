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

	var sine1 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    60 * time.Second,
			Amplitude: 100,
			Offset:    0,
			DutyCycle: 1,
			Variable:  &cpu,
		},
	}

	var sine2 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    60 * time.Second,
			Amplitude: 60,
			Offset:    4,
			DutyCycle: 1,
			Variable:  &ram,
		},
	}

	var sine3 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    70 * time.Second,
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

	//Redis benchmark
	client := v.Variable{Name: "client"}
	size := v.Variable{Name: "size"}
	keyset := v.Variable{Name: "keyset"}
	pipeline := v.Variable{Name: "pipeline"}

	var redissine1 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    50 * time.Second,
			Amplitude: 250,
			Offset:    1,
			DutyCycle: 1,
			Variable:  &client,
		},
	}

	var redissine2 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    15 * time.Second,
			Amplitude: 10,
			Offset:    1,
			DutyCycle: 1,
			Variable:  &size,
		},
	}

	var redissine3 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    70 * time.Second,
			Amplitude: 1000,
			Offset:    10,
			DutyCycle: 1,
			Variable:  &keyset,
		},
	}

	var redissine4 signal.Signal = signal.WaveletFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    60 * time.Second,
			Amplitude: 100,
			Offset:    10,
			DutyCycle: 1,
			Variable:  &pipeline,
		},
	}

	var redisBench actions.IAction = &actions.RedisBench{Test: "ping,set,get"}
	redisBench.Default()
	redisBench.AddTarget("MyRedisContainer")
	redisBench.AddVariable(&client)
	redisBench.AddVariable(&size)
	redisBench.AddVariable(&keyset)
	redisBench.AddVariable(&pipeline)

	var manager = signal.NewManager("main")
	manager.AddMonkey(signal.NewChaosMonkey("sampler1", append([]*signal.Signal{}, &sine1), 0*time.Second, 1*time.Second, stressContainer))
	manager.AddMonkey(signal.NewChaosMonkey("sampler2", append([]*signal.Signal{}, &sine2), 25*time.Second, 1*time.Second, stressContainer2))
	manager.AddMonkey(signal.NewChaosMonkey("sampler3", append([]*signal.Signal{}, &sine3), 50*time.Second, 1*time.Second, stressContainer3))

	manager.AddMonkey(signal.NewChaosMonkey("redisbench", append([]*signal.Signal{}, &redissine1, &redissine2, &redissine3, &redissine4), 0*time.Second, 1*time.Second, redisBench))

	manager.GenerateScripts()

	fmt.Println("Done!")
}
