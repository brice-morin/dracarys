package docker

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"

	"../actions"
)

type DockerCLI struct {
  docker *client.Client
}

func (c *DockerCLI) NewDockerCLI() {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	c.docker = cli
}

/*Kills v random containers*/
type KillContainer struct {
	actions.Action
	DockerCLI
}

func (a KillContainer) Do(v int64) {
	if (a.docker == nil) {
		a.NewDockerCLI()
	}

  containers, err := a.docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
	  r  := rand.New(rand.NewSource(time.Now().UnixNano()))

	  var i int64 = 0
	  for i < v {
			id := r.Int() % len(containers) //FIXME: ensure that we get different IDs (currently, could be n times the same)
			fmt.Println("Killing container ", containers[id].ID[:12])
			err := a.docker.ContainerKill(context.Background(), containers[id].ID[:12], "SIGKILL")
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
	if (a.docker == nil) {
		a.NewDockerCLI()
	}

  containers, err := a.docker.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
	  r  := rand.New(rand.NewSource(time.Now().UnixNano()))
		duration := 2*time.Second
	  var i int64 = 0
	  for i < v {
			id := r.Int() % len(containers) //FIXME: ensure that we get different IDs (currently, could be n times the same)
			fmt.Println("Restarting container ", containers[id].ID[:12])
			err := a.docker.ContainerRestart(context.Background(), containers[id].ID[:12], &duration)
			if err != nil {
				fmt.Println("ERROR: ", err)
			}
		  i = i + 1
	  }
  }
}

//TODO: Actions to rollout updates e.g. change image, change available resources
