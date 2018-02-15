package actions

import (
	"fmt"
	"strings"
)

var (
	Debug bool
)

/*Limits quota allocated to a container*/
type LimitContainer struct {
	Action
	AsInt bool
}

func (a *LimitContainer) Default() IAction {
	a.SetScope(ALL)
	a.SetType(CONTAINER)
	return a
}

func (a *LimitContainer) Print() string {
	v := a.GetVariables()[0]
	name := strings.Replace(v.Name, "-", "_", -1)
	return fmt.Sprintf("docker update --%s ${%s}%s $target", v.Name, name, v.Unit)
}

/*Stresses a container by generating CPU, RAM, IO and HDD activities for v seconds*/
type StressContainer struct {
	Action
	Resource string //cpu, vm, io or hdd
}

func (a *StressContainer) Default() IAction {
	a.SetScope(ALL)
	a.SetType(CONTAINER)
	return a
}

func (a *StressContainer) Print() string {
	v := a.GetVariables()[0]
	return fmt.Sprintf("docker exec -d $target stress --%s 2 --timeout ${%s}s --verbose", v.Name, v.Name)
}

/*Limit network rate to  v kbit/s for traffic  on all containers*/
type NetworkRate struct {
	Action
}

func (a *NetworkRate) Default() IAction {
	a.SetScope(ALL)
	a.SetType(CONTAINER)
	return a
}

func (a *NetworkRate) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 rate --rate ${sample}kbit $container", int64(a.GetTick().Seconds()))
}

/*Delay packet on all containers by v ms*/
type PacketDelay struct {
	Action
}

func (a *PacketDelay) Default() IAction {
	a.SetScope(ALL)
	a.SetType(CONTAINER)
	return a
}

func (a *PacketDelay) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 delay --time ${sample}ms $container", int64(a.GetTick().Seconds()))
}

/*Drops v% of incoming packet on all containers*/
type PacketLoss struct {
	Action
}

func (a *PacketLoss) Default() IAction {
	a.SetScope(ALL)
	a.SetType(CONTAINER)
	return a
}

func (a *PacketLoss) Print() string {
	return fmt.Sprintf("pumba netem --duration %ds --interface eth0 loss --percent ${sample} 25 $container", int64(a.GetTick().Seconds()))
}

/*Kills v random containers*/
type KillContainer struct {
	Action
}

func (a *KillContainer) Default() IAction {
	a.SetScope(RND)
	a.SetType(CONTAINER)
	return a
}

func (a *KillContainer) Print() string {
	return "docker rm -f $container"
}

/*Restarts v random containers*/
type RestartContainer struct {
	Action
}

func (a *RestartContainer) Default() IAction {
	a.SetScope(RND)
	a.SetType(CONTAINER)
	return a
}

func (a *RestartContainer) Print() string {
	return "docker restart $container"
}

/*Scale a random service to v instances*/
type ScaleService struct {
	Action
}

func (a *ScaleService) Default() IAction {
	a.SetScope(RND)
	a.SetType(SERVICE)
	return a
}

func (a *ScaleService) Print() string {
	return "docker service scale $service=$sample"
}

/*Disconnect a random container from all its networks for v seconds*/
type NetworkDisconnect struct {
	Action
}

func (a *NetworkDisconnect) Default() IAction {
	a.SetScope(RND)
	a.SetType(CONTAINER)
	return a
}

func (a *NetworkDisconnect) PrintHelper() string {
	return "#$1 container, $2 duration\nfunction _disconnect {\n" +
		"  container=$1\n" +
		"  duration=$2\n" +
		"  mapfile -t networks < <(docker inspect --format='{{range .NetworkSettings.Networks}}{{.NetworkID}}{{end}}' $container)\n" +
		"  for (( i=0; i<${#networks[@]}; i++ )); do\n" +
		"    (docker network disconnect ${networks[i]} $container\n" +
		"    sleep ${duration}s\n" +
		"    docker network connect ${networks[i]} $container)&\n" +
		"  done\n" +
		"}\n"
}

func (a *NetworkDisconnect) Print() string {
	return "_disconnect $container $sample"
}

//TODO: Actions to rollout updates e.g. change image, change available resources
