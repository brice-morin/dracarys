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

	cpu_stress := v.Variable{Name: "c"}
	ram_stress := v.Variable{Name: "m"}
	io_stress := v.Variable{Name: "i"}
	hdd_stress := v.Variable{Name: "d"}

	var sineA signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    457 * time.Second,
			Amplitude: 30,
			Offset:    0,
			DutyCycle: 1,
			Variable:  &cpu_stress,
		},
	}

	var sineB signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    419 * time.Second,
			Amplitude: 30,
			Offset:    0,
			DutyCycle: 1,
			Variable:  &ram_stress,
		},
	}

	var sineC signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    347 * time.Second,
			Amplitude: 30,
			Offset:    0,
			DutyCycle: 1,
			Variable:  &io_stress,
		},
	}

	var sineD signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    523 * time.Second,
			Amplitude: 30,
			Offset:    0,
			DutyCycle: 1,
			Variable:  &hdd_stress,
		},
	}

	var stressContainer actions.IAction = &actions.StressContainer{}
	stressContainer.Default()
	stressContainer.AddVariable(&cpu_stress)
	stressContainer.SetScope(actions.NEW)
	var stressContainer2 actions.IAction = &actions.StressContainer{}
	stressContainer2.Default()
	stressContainer2.AddVariable(&ram_stress)
	stressContainer2.SetScope(actions.NEW)
	var stressContainer3 actions.IAction = &actions.StressContainer{}
	stressContainer3.Default()
	stressContainer3.AddVariable(&io_stress)
	stressContainer3.SetScope(actions.NEW)
	var stressContainer4 actions.IAction = &actions.StressContainer{}
	stressContainer4.Default()
	stressContainer4.AddVariable(&hdd_stress)
	stressContainer4.SetScope(actions.NEW)

	cpu := v.Variable{Name: "cpus"}
	ram := v.Variable{Name: "memory", Unit: "m"}
	//io := v.Variable{Name: "blkio-weight"}

	var sine1 signal.Signal = signal.SineFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    271 * time.Second,
			Amplitude: 0.95,
			Offset:    0.05,
			DutyCycle: 1,
			Variable:  &cpu,
		},
	}

	var sine2 signal.Signal = signal.SineFloat{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    307 * time.Second,
			Amplitude: 60,
			Offset:    4,
			DutyCycle: 1,
			Variable:  &ram,
		},
	}

	/*var sine3 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    113 * time.Second,
			Amplitude: 990,
			Offset:    10,
			DutyCycle: 1,
			Variable:  &io,
		},
	}*/

	fmt.Println("Starting...")

	var limitContainer actions.IAction = &actions.LimitContainer{}
	limitContainer.Default()
	limitContainer.AddVariable(&cpu)
	var limitContainer2 actions.IAction = &actions.LimitContainer{}
	limitContainer2.Default()
	limitContainer2.AddVariable(&ram)
	/*var limitContainer3 actions.IAction = &actions.LimitContainer{AsInt: true}
	limitContainer3.Default()
	limitContainer3.AddVariable(&io)*/

	//Redis benchmark
	client := v.Variable{Name: "client"}
	size := v.Variable{Name: "size"}
	keyset := v.Variable{Name: "keyset"}
	pipeline := v.Variable{Name: "pipeline"}

	var redissine1 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    101 * time.Second,
			Amplitude: 250,
			Offset:    1,
			DutyCycle: 1,
			Variable:  &client,
		},
	}

	var redissine2 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    157 * time.Second,
			Amplitude: 10,
			Offset:    1,
			DutyCycle: 1,
			Variable:  &size,
		},
	}

	var redissine3 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    233 * time.Second,
			Amplitude: 1000,
			Offset:    10,
			DutyCycle: 1,
			Variable:  &keyset,
		},
	}

	var redissine4 signal.Signal = signal.Sine{ //Stress CPU, RAM, HDD, IO, with phase shift
		PeriodicSignal: signal.PeriodicSignal{
			Period:    283 * time.Second,
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
	manager.AddMonkey(signal.NewChaosMonkey("stress1", append([]*signal.Signal{}, &sineA), 0*time.Second, 30*time.Second, stressContainer))
	manager.AddMonkey(signal.NewChaosMonkey("stress2", append([]*signal.Signal{}, &sineB), 53*time.Second, 30*time.Second, stressContainer2))
	manager.AddMonkey(signal.NewChaosMonkey("stress3", append([]*signal.Signal{}, &sineC), 67*time.Second, 30*time.Second, stressContainer3))
	manager.AddMonkey(signal.NewChaosMonkey("stress4", append([]*signal.Signal{}, &sineD), 79*time.Second, 30*time.Second, stressContainer4))

	manager.AddMonkey(signal.NewChaosMonkey("limit1", append([]*signal.Signal{}, &sine1), 0*time.Second, 30*time.Second, limitContainer))
	manager.AddMonkey(signal.NewChaosMonkey("limit2", append([]*signal.Signal{}, &sine2), 103*time.Second, 30*time.Second, limitContainer2))
	//manager.AddMonkey(signal.NewChaosMonkey("limit3", append([]*signal.Signal{}, &sine3), 0*time.Second, 4*time.Second, limitContainer3))

	manager.AddMonkey(signal.NewChaosMonkey("redisbench", append([]*signal.Signal{}, &redissine1, &redissine2, &redissine3, &redissine4), 0*time.Second, 30*time.Second, redisBench))

	manager.GenerateScripts()

	fmt.Println("Done!")
}
