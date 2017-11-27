package docker

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os/exec"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	//"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	pumba "github.com/brice-morin/pumba/action"
	container "github.com/brice-morin/pumba/container"

	"../actions"
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

//FIXME: we should reuse the function provided by Pumba (e.g. so as to implement any new action requiring to exec into containers)
/*func (c *DockerCLI) ExecIntoContainer(ctx context.Context, cont types.Container, execCmd string, execArgs []string, privileged bool, retryOnError bool) error {
	// trim all spaces from cmd
	execCmd = strings.Replace(execCmd, " ", "", -1)
	exitCode := 1
	doContinue := retryOnError

	config := types.ExecConfig{
		Privileged: privileged,
		Cmd:        append([]string{execCmd}, execArgs...),
	}
	for doContinue { //FIXME: this is mostly a hack... e.g. we need to understand why iptables commands sometimes work, sometimes not...
		// execute the command
		exec, err := docker.ContainerExecCreate(ctx, cont.ID, config)
		if err != nil {
			return err
		}
		fmt.Println("Executing", execCmd, execArgs)
		//log.Debugf("Starting Exec %s %s (%s)", execCmd, execArgs, exec.ID)
		err = docker.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
		if err != nil {
			return err
		}
		exitInspect, err := docker.ContainerExecInspect(ctx, exec.ID)
		if err != nil {
			fmt.Println(err)
		}
		if exitInspect.ExitCode != 0 {
			fmt.Println("ERROR: Command[", execCmd, execArgs, "] returned", exitInspect.ExitCode)
			//return fmt.Errorf("command '%s' failed in %s (%s) container; run it in manually to debug", execCmd, c.Name(), c.ID())
		} else {
			fmt.Println("SUCCESS: Command[", execCmd, execArgs, "] returned", exitInspect.ExitCode)
		}
		exitCode = exitInspect.ExitCode
		if !retryOnError {
			doContinue = false
		} else {
			doContinue = exitCode == 1 //we stop e.g. if we get 127 or 126
		}
	}
	return nil
}*/

/*Limit network rate to  v kbit/s for traffic  on all containers*/
type NetworkRate struct {
	actions.Action
	DockerCLI
	PumbaCLI
}

func (a NetworkRate) Do(v int64) {
	a.NewDockerCLI()
	a.NewPumbaCLI()

	a.DockerCLI.mux.Lock()
	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	a.DockerCLI.mux.Unlock()
	if err != nil {
		panic(err)
	}

	cmd := pumba.CommandNetemRate{
		NetInterface: "eth0",
		IPs:          []net.IP{net.IPv4(255, 255, 255, 255)},
		Duration:     a.Action.Duration,
		Rate:         fmt.Sprintf("%dkbit", v),
		Image:        "",
	}

	for _, container := range containers {
		a.DockerCLI.mux.Lock()
		info, _ := docker.ContainerInspect(context.Background(), container.ID)
		a.DockerCLI.mux.Unlock()
		if container.Labels["protected"] != "true" {
			go func() {
				fmt.Println("Container", info.Name, "will have limited network rate down to", cmd.Rate, "kbit/s for", cmd.Duration)
				a.PumbaCLI.mux.Lock()
				err := pumbaChaos.NetemRateContainers(context.Background(), pumbaClient, []string{info.Name}, "", cmd)
				a.PumbaCLI.mux.Unlock()
				if err != nil {
					fmt.Println("Error:", err)
				}
			}()
		} else {
			fmt.Println("container", info.Name, "is protected")
		}
	}
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketDelay struct {
	actions.Action
	DockerCLI
	PumbaCLI
}

func (a PacketDelay) Do(v int64) {
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
			go func() {
				fmt.Println("Container", info.Name, "will delay packets with", cmd.Time, "with jitter", cmd.Jitter, "for", cmd.Duration)
				a.PumbaCLI.mux.Lock()
				err := pumbaChaos.NetemDelayContainers(context.Background(), pumbaClient, []string{info.Name}, "", cmd)
				a.PumbaCLI.mux.Unlock()
				if err != nil {
					fmt.Println("Error:", err)
				}
			}()
		} else {
			fmt.Println("container", info.Name, "is protected")
		}
	}
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketLoss struct {
	actions.Action
	DockerCLI
	PumbaCLI
}

func (a PacketLoss) Do(v int64) {
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
			go func() {
				fmt.Println("Container", info.Name, "will drop", cmd.Percent, "percent of packets for", cmd.Duration)
				a.PumbaCLI.mux.Lock()
				err := pumbaChaos.NetemLossRandomContainers(context.Background(), pumbaClient, []string{info.Name}, "", cmd)
				a.PumbaCLI.mux.Unlock()
				if err != nil {
					fmt.Println("Error:", err)
				}
			}()
		} else {
			fmt.Println("container", info.Name, "is protected")
		}
	}
}

/*Kills v random containers*/
type KillContainer struct {
	actions.Action
	DockerCLI
}

func (a KillContainer) Do(v int64) { //TODO: make sure we do not kill containers with label "protected:true"
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
			fmt.Println("Killing container ", containers[id].ID[:12])
			err := docker.ContainerKill(context.Background(), containers[id].ID[:12], "SIGKILL")
			if err != nil {
				fmt.Println("ERROR: ", err)
			}
			i = i + 1
		}
	}
}

/*Restarts v random containers*/
type RestartContainer struct {
	actions.Action
	DockerCLI
}

func (a RestartContainer) Do(v int64) {
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
	actions.Action
	DockerCLI
}

func (a ScaleService) Do(v int64) {
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

		cmd := exec.Command("docker", "service", "scale", fmt.Sprintf("%s=%d", name, v))
		var out bytes.Buffer
		cmd.Stdout = &out
		err := cmd.Run()
		if err != nil {
			fmt.Println("ERROR: ", err)
		}
		fmt.Printf("debug: %q\n", out.String())
	}
}

//TODO: Actions to rollout updates e.g. change image, change available resources
