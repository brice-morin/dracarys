package docker

import (
	"fmt"
	"math/rand"
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
)

type DockerCLI struct{} //FIXME: we could only use PumbaCLI, which also instantiate a DockerCLI

type PumbaCLI struct{}

func (c *DockerCLI) NewDockerCLI() { //FIXME: we could only use PumbaCLI, which also instantiate a DockerCLI
	if docker == nil {
		cli, err := client.NewEnvClient()
		if err != nil {
			panic(err)
		}
		docker = cli
	}
}

func (c *PumbaCLI) NewPumbaCLI() {
	if pumbaChaos == nil {
		pumbaChaos = pumba.NewChaos()
	}
	if pumbaClient == nil {
		pumbaClient = container.NewEnvClient()
	}
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

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketLoss struct {
	actions.Action
	DockerCLI
	PumbaCLI
}

func (a PacketLoss) Do(v int64) {
	a.NewDockerCLI()
	a.NewPumbaCLI()

	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	cmd := pumba.CommandNetemLossRandom{
		NetInterface: "eth0",
		IPs:          nil,
		Duration:     1 * time.Millisecond,
		Percent:      11.5,
		Correlation:  25.53,
	}

	for _, container := range containers {
		if container.Labels["protected"] != "true" {
			err := pumbaChaos.NetemLossRandomContainers(context.Background(), pumbaClient, []string{container.Names[0]}, "", cmd)
			if err != nil {
				fmt.Println("Error:", err)
			}
		} else {
			fmt.Println("container", container.Names[0], "is protected")
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

	containers, err := docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		duration := 2 * time.Second
		var i int64 = 0
		for i < v {
			id := r.Int() % len(containers) //FIXME: ensure that we get different IDs (currently, could be n times the same)
			fmt.Println("Restarting container ", containers[id].ID[:12])
			err := docker.ContainerRestart(context.Background(), containers[id].ID[:12], &duration)
			if err != nil {
				fmt.Println("ERROR: ", err)
			}
			i = i + 1
		}
	}
}

//TODO: Actions to rollout updates e.g. change image, change available resources
