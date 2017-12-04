package actions

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	pumba "github.com/brice-morin/pumba/action"
	container "github.com/brice-morin/pumba/container"
)

var (
	docker      *client.Client //FIXME: we could only use PumbaCLI, which also instantiate a DockerCLI
	pumbaChaos  pumba.Chaos
	pumbaClient container.Client
	Debug       bool
)

type DockerCLI struct {
	mux sync.Mutex
} //FIXME: we could only use PumbaCLI, which also instantiate a DockerCLI

type PumbaCLI struct {
	mux sync.Mutex
}

func (c *DockerCLI) NewDockerCLI() { //FIXME: we could only use PumbaCLI, which also instantiate a DockerCLI
	c.mux.Lock()
	if docker == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		if Debug {
			events, errors := cli.Events(context.Background(), types.EventsOptions{})
			go func() {
				for {
					select {
					case ev := <-events:
						fmt.Println("DOCKER-INFO:", ev)
					case err := <-errors:
						fmt.Println("DOCKER-ERRO:", err)
					}
				}
			}()
		}
		docker = cli
	}
	c.mux.Unlock()
}

func (c *PumbaCLI) NewPumbaCLI() {
	c.mux.Lock()
	if pumbaChaos == nil {
		pumbaChaos = pumba.NewChaos()
	}
	if pumbaClient == nil {
		pumbaClient = container.NewEnvClient()
	}
	c.mux.Unlock()
}

func Exec(execCmd string, args ...string) int {
	//fmt.Println("DEEEEEEEEEEBUUUUUUUUUUUG:", execCmd, args)
	cmd := exec.Command(execCmd, args...)
	cmd.Env = os.Environ()
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		//fmt.Println("ERROR: ", err)
		return 1
	}
	return 0
}

//FIXME: we should reuse the function provided by Pumba (e.g. so as to implement any new action requiring to exec into containers)
func (c *DockerCLI) ExecIntoContainer(ctx context.Context, id string, execCmd string, args ...string) {
	retries := 0
	success := false

	for !(success || retries == 10) { //FIXME: for some reasons, some commands fail without any good explanation... but eventually succeed...
		exit := Exec("docker", append([]string{"exec", execCmd}, args...)...)
		if exit == 0 {
			success = true
			break
		}
		retries = retries + 1
	}
	if !success {
		fmt.Println("Command", execCmd, args, "failed for container", id)
	}
	/*
		// prepare exec config
		config := types.ExecConfig{
			User:       "root",
			Privileged: true,
			Cmd:        append([]string{execCmd}, args...),
		}
		// execute the command
		for !(success || retries == 10) { //FIXME: for some reasons, some commands fail without any good explanation... but eventually succeed...
			c.mux.Lock()
			exec, err := docker.ContainerExecCreate(ctx, id, config)
			if err != nil {
				fmt.Println("ContainerExecCreate [ERROR]", execCmd, ":", err)
			}
			err = docker.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
			if err != nil {
				fmt.Println("ContainerExecStart [ERROR]", execCmd, ":", err)
			}
			exitInspect, err := docker.ContainerExecInspect(ctx, exec.ID)
			if err != nil {
				fmt.Println("ContainerExecInspect [ERROR]", execCmd, ":", err)
			}
			c.mux.Unlock()
			//fmt.Println("exitInspect", exitInspect)
			if exitInspect.ExitCode != 0 {
				//fmt.Println("ExitCode[", retries, "]", execCmd, ":", exitInspect.ExitCode, exitInspect.Running)
			} else {
				//fmt.Println("Success after", retries, "retries")
				success = true
				break
			}
			retries = retries + 1
			time.Sleep(250 * time.Microsecond)
		}
		if !success {
			fmt.Println("Command", config.Cmd, "failed for container", id)
		}
	*/
}

type DockerAction struct {
	Action
	DockerCLI
}

func (a *DockerAction) _do(v int64, do func(int64, string)) {
	a.NewDockerCLI()
	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		fmt.Println(err)
		//panic(err)
		return
	}
	for _, container := range containers {
		if container.Labels["protected"] != "true" {
			do(v, container.ID)
		}
	}
}

func (a *DockerAction) DoTo(v int64, id string) {}

/*Stresses a container by generating CPU, RAM, IO and HDD activities for v seconds*/
type StressContainer struct {
	DockerAction
	Resource string //cpu, vm, io or hdd
}

func (a *StressContainer) _do(v int64, do func(int64, string)) {
	a.DockerAction._do(v, a.DoTo)
}

func (a *StressContainer) Do(v int64) {
	fmt.Println("stressing", a.Resource, fmt.Sprintf("for %ds", v))
	a._do(v, a.DoTo)
}

func (a *StressContainer) DoTo(v int64, id string) {
	fmt.Println("stressing", a.Resource, "on container", id, fmt.Sprintf("for %ds", v))
	if v > 0 {
		a.DockerCLI.ExecIntoContainer(context.Background(), id, "stress", fmt.Sprintf("--%s 16 --timeout %ds --verbose", a.Resource, v))
	}
}

func (a *StressContainer) Print() string {
	return fmt.Sprintf("docker exec -d $container stress --%s 16 --timeout ${sample}s --verbose", a.Resource)
}

/*Limit network rate to  v kbit/s for traffic  on all containers*/
type NetworkRate struct {
	DockerAction
	PumbaCLI
}

func (a *NetworkRate) _do(v int64, do func(int64, string)) {
	a.DockerAction._do(v, a.DoTo)
}

/*...*/
func (a *NetworkRate) DoTo(v int64, id string) {
	a.NewDockerCLI()
	a.NewPumbaCLI()

	cmd := pumba.CommandNetemRate{
		NetInterface: "eth0",
		IPs:          []net.IP{net.IPv4(255, 255, 255, 255)},
		Duration:     a.Action.Duration,
		Rate:         fmt.Sprintf("%dbit", v),
		Image:        "",
	}

	go func() {
		fmt.Println("Container", id, "will have limited network rate down to", cmd.Rate, "bit/s for", cmd.Duration)
		//a.PumbaCLI.mux.Lock()
		err := pumbaChaos.NetemRateContainers(context.Background(), pumbaClient, []string{id}, "", cmd)
		//a.PumbaCLI.mux.Unlock()
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()

}

func (a *NetworkRate) Do(v int64) {
	fmt.Println("Limiting network rate to", fmt.Sprintf("%dbit/s", v))
	a._do(v, a.DoTo)
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketDelay struct {
	DockerAction
	PumbaCLI
}

func (a *PacketDelay) Do(v int64) {
	a.NewDockerCLI()
	a.NewPumbaCLI()

	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	cmd := pumba.CommandNetemDelay{
		NetInterface: "eth0",
		IPs:          []net.IP{net.IPv4(255, 255, 255, 255)},
		Duration:     a.Action.Duration,
		Time:         int(v),
		Jitter:       int(v / 10),
		Correlation:  25,
	}

	for _, container := range containers {
		a.DockerCLI.mux.Lock()
		info, _ := docker.ContainerInspect(context.Background(), container.ID)
		a.DockerCLI.mux.Unlock()
		if container.Labels["protected"] != "true" {
			//go func() {
			fmt.Println("Container", info.Name, "will delay packets with", cmd.Time, "with jitter", cmd.Jitter, "for", cmd.Duration)
			a.PumbaCLI.mux.Lock()
			err := pumbaChaos.NetemDelayContainers(context.Background(), pumbaClient, []string{info.Name}, "", cmd)
			a.PumbaCLI.mux.Unlock()
			if err != nil {
				fmt.Println("Error:", err)
			}
			//}()
		} else {
			fmt.Println("container", info.Name, "is protected")
		}
	}
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketLoss struct {
	DockerAction
	PumbaCLI
}

func (a *PacketLoss) Do(v int64) {
	fmt.Println("DEBUG")
	a.NewDockerCLI()
	a.NewPumbaCLI()

	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	cmd := pumba.CommandNetemLossRandom{
		NetInterface: "eth0",
		IPs:          []net.IP{net.IPv4(255, 255, 255, 255)},
		Duration:     a.Action.Duration,
		Percent:      float64(v),
		Correlation:  25,
	}

	for _, container := range containers {
		a.DockerCLI.mux.Lock()
		info, _ := docker.ContainerInspect(context.Background(), container.ID)
		a.DockerCLI.mux.Unlock()
		if container.Labels["protected"] != "true" {
			//go func() {
			fmt.Println("Container", info.Name, "will drop", cmd.Percent, "percent of packets for", cmd.Duration)
			a.PumbaCLI.mux.Lock()
			err := pumbaChaos.NetemLossRandomContainers(context.Background(), pumbaClient, []string{info.Name}, "", cmd)
			a.PumbaCLI.mux.Unlock()
			if err != nil {
				fmt.Println("Error:", err)
			}
			//}()
		} else {
			fmt.Println("container", info.Name, "is protected")
		}
	}
}

/*Kills v random containers*/
type KillContainer struct {
	DockerAction
}

func (a *KillContainer) Do(v int64) { //TODO: make sure we do not kill containers with label "protected:true"
	if docker == nil {
		a.NewDockerCLI()
	}

	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		var i int64 = 0
		for i < v {
			id := r.Int() % len(containers) //FIXME: ensure that we get different IDs (currently, could be n times the same)
			if containers[id].Labels["protected"] == "true" {
				continue
			} else {
				fmt.Println("Killing container ", containers[id].ID[:12])
				err := docker.ContainerKill(context.Background(), containers[id].ID[:12], "SIGKILL")
				if err != nil {
					fmt.Println("ERROR: ", err)
				}
				i = i + 1
			}
		}
	}
}

/*Restarts v random containers*/
type RestartContainer struct {
	DockerAction
}

func (a *RestartContainer) Do(v int64) {
	if docker == nil {
		a.NewDockerCLI()
	}

	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		duration := 2 * time.Second
		var i int64 = 0
		for i < v {
			id := r.Int() % len(containers)                   //FIXME: ensure that we get different IDs (currently, could be n times the same)
			if containers[id].Labels["protected"] == "true" { //Note: there is a tiny chance it will loop forever, if random keeps on choosing protected containers...
				continue
			} else {
				fmt.Println("Restarting container ", containers[id].ID[:12])
				a.DockerCLI.mux.Lock()
				err := docker.ContainerRestart(context.Background(), containers[id].ID[:12], &duration)
				a.DockerCLI.mux.Unlock()
				if err != nil {
					fmt.Println("ERROR: ", err)
				}
				i = i + 1
			}
		}
	}
}

/*Scale a service with v up or down (v could be positive or negative)*/
type ScaleService struct {
	DockerAction
}

func (a *ScaleService) Do(v int64) {
	if docker == nil {
		a.NewDockerCLI()
	}

	a.DockerCLI.mux.Lock()
	services, err := docker.ServiceList(context.Background(), types.ServiceListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	if len(services) > 0 {
		var id int
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for { //Find one random service that is not protected
			id = r.Int() % len(services)
			if services[id].Spec.TaskTemplate.ContainerSpec.Labels["protected"] != "true" {
				break
			}
		}

		name := services[id].Spec.Name
		fmt.Println("Scaling service ", name)

		//FIXME: figure out how to do it with the SDK. We just use os/exec meanwhile...
		Exec("docker", "service", "scale", fmt.Sprintf("%s=%d", name, v))
	}
}

/*Disconnect a random container from network for v seconds*/
type NetworkDisconnect struct {
	DockerAction
}

func (a *NetworkDisconnect) Do(v int64) {
	if docker == nil {
		a.NewDockerCLI()
	}

	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 && v > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for {
			id := r.Int() % len(containers)                   //FIXME: ensure that we get different IDs (currently, could be n times the same)
			if containers[id].Labels["protected"] == "true" { //Note: there is a tiny chance it will loop forever, if random keeps on choosing protected containers...
				continue
			} else {
				//go func() {
				a.DockerCLI.mux.Lock()
				duration, _ := time.ParseDuration(fmt.Sprintf("%ds", v))
				for net := range containers[id].NetworkSettings.Networks {
					fmt.Println("Disconnecting container ", containers[id].ID[:12], "from network", containers[id].NetworkSettings.Networks[net].NetworkID, "for", duration)
					err := docker.NetworkDisconnect(context.Background(), containers[id].NetworkSettings.Networks[net].NetworkID, containers[id].ID, true)
					if err != nil {
						fmt.Println("ERROR: ", err)
					}
				}

				time.Sleep(duration)

				for net := range containers[id].NetworkSettings.Networks {
					fmt.Println("Reconnecting container ", containers[id].ID[:12], "to network", containers[id].NetworkSettings.Networks[net].NetworkID)
					err := docker.NetworkConnect(context.Background(), containers[id].NetworkSettings.Networks[net].NetworkID, containers[id].ID, &network.EndpointSettings{})
					if err != nil {
						fmt.Println("ERROR: ", err)
					}
				}

				a.DockerCLI.mux.Unlock()
				//}()
				break
			}
		}
	}
}

//TODO: Actions to rollout updates e.g. change image, change available resources
