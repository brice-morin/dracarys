package actions

import (
	"fmt"
)

var (
	Debug bool
)

/*Stresses a container by generating CPU, RAM, IO and HDD activities for v seconds*/
type StressContainer struct {
	Action
	Resource string //cpu, vm, io or hdd
}

func (a *StressContainer) Print() string {
	return fmt.Sprintf("docker exec -d $container stress --%s 16 --timeout ${sample}s --verbose", a.Resource)
}

/*Limit network rate to  v kbit/s for traffic  on all containers*/
type NetworkRate struct {
	Action
}

func (a *NetworkRate) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 rate --rate ${sample}kbit $container", int64(a.GetTick().Seconds()))
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketDelay struct {
	Action
}

func (a *PacketDelay) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 delay --time ${sample}ms $container", int64(a.GetTick().Seconds()))
}

/*Drops v% of incoming packet on all containers*/ //TODO: make an action targeting specific/random sub-sets of containers
type PacketLoss struct {
	Action
}

func (a *PacketLoss) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 loss --percent ${sample} 25 $container", int64(a.GetTick().Seconds()))
}

/*Kills v random containers*/
type KillContainer struct {
	Action
}

func (a *KillContainer) Print() string {
	return "docker rm -f $container"
}

/*Restarts v random containers*/
type RestartContainer struct {
	Action
}

func (a *RestartContainer) Print() string {
	return "docker restart $container"
}

/*Scale a service with v up or down (v could be positive or negative)*/
/*type ScaleService struct {
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
*/

/*Disconnect a random container from network for v seconds*/
/*type NetworkDisconnect struct {
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
}*/

//TODO: Actions to rollout updates e.g. change image, change available resources
