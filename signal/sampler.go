package signal

import (
	"bytes"
	"container/ring"
	"fmt"
	"os"
	"time"

	"../actions"
)

type Sample struct {
	t int64
	v int64
}

//TODO: add the possibility to have several signals, typically one for #containers to impact, and one for duration
type ChaosMonkey struct {
	Name    string
	samples *ring.Ring
	Signal
	Shift  time.Duration
	Rate   time.Duration
	Action actions.IAction
	Ticker *time.Ticker
}

func NewChaosMonkey(name string, sig *Signal, shift time.Duration, r time.Duration, a actions.IAction) *ChaosMonkey {
	a.SetTick(r)
	var sampler = ChaosMonkey{
		Name:   name,
		Signal: *sig,
		Rate:   r,
		Action: a,
	}
	return &sampler
}

func printResources(a actions.IAction) string {
	var buffer bytes.Buffer
	for i, resource := range a.GetResources() {
		if i > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("\"" + resource + "\"")
	}
	return buffer.String()
}

func (s *ChaosMonkey) GenerateScript() {
	var buffer bytes.Buffer
	buffer.WriteString("#! /bin/bash\n")
	buffer.WriteString("alive=true\n")
	buffer.WriteString("function _term {\n")
	buffer.WriteString("  alive=false\n")
	buffer.WriteString("  echo \"Waiting " + fmt.Sprintf("%ds", int64(s.Rate.Seconds())) + "for processes to terminate... Ctrl+C again to quit now.\"\n")
	buffer.WriteString(fmt.Sprintf("  sleep %ds\n", int64(s.Rate.Seconds())))
	buffer.WriteString("  for job in `jobs -p`\n")
	buffer.WriteString("  do\n")
	buffer.WriteString("    #echo $job\n")
	buffer.WriteString("    wait $job\n")
	buffer.WriteString("  done\n")
	//TODO: add cleaning operations (restablishing containers in a decent state when we will kill the script)
	//NOTE: though ideally, operations will leave the container in its initial state after they finish, so stopping the while loop and waiting for processes should be OK...
	buffer.WriteString("  echo \"The End.\"\n")
	buffer.WriteString("  exit\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("trap _term INT\n\n")

	buffer.WriteString(s.Action.PrintHelper() + "\n")

	buffer.WriteString("declare -a samples=(")
	for i := 0; i < int(s.Signal.GetPeriod().Seconds()); i = i + int(s.Rate.Seconds()) {
		if i > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString(fmt.Sprintf("%d", s.Signal.Sample(int64(i+int(s.Shift.Seconds())))))
	}
	buffer.WriteString(")\n")

	if s.Action.GetScope() == actions.SOME {
		switch s.Action.GetTarget() {
		case actions.CONTAINER:
			buffer.WriteString("declare -a containers=(" + printResources(s.Action) + ")\n")
		case actions.NETWORK:
			buffer.WriteString("declare -a networks=(" + printResources(s.Action) + ")\n")
		case actions.SERVICE:
			buffer.WriteString("declare -a services=(" + printResources(s.Action) + ")\n")
		}
	}

	buffer.WriteString("echo \"--------\" >> " + s.Name + ".log\n")
	buffer.WriteString("echo \"--------\" >> " + s.Name + ".sig\n")

	buffer.WriteString("while $alive; do\n")
	buffer.WriteString("  for sample in \"${samples[@]}\"; do\n")
	buffer.WriteString("    #echo $sample\n")
	inputs := "containers"
	input := "container"
	switch s.Action.GetTarget() {
	case actions.CONTAINER:
		if s.Action.GetScope() != actions.SOME {
			buffer.WriteString("    mapfile -t containers < <(docker ps -q --format \"{{.Names}}\" --filter label=eu.stamp.dracarys=true)\n") //TODO --filter "label=dracarys=true"
		}
	case actions.NETWORK:
		if s.Action.GetScope() != actions.SOME {
			buffer.WriteString("    mapfile -t networks < <(docker network ls -q --format \"{{.Name}}\")\n") //TODO --filter "label=dracarys=true"
		}
		inputs = "networks"
		input = "network"
	case actions.SERVICE:
		if s.Action.GetScope() != actions.SOME {
			buffer.WriteString("    mapfile -t services < <(docker service ls -q --format \"{{.Name}}\" --filter label=eu.stamp.dracarys=true)\n") //TODO --filter "label=dracarys=true"
		}
		inputs = "services"
		input = "service"
	}

	buffer.WriteString("    if [[ ${#" + inputs + "[@]} -eq \"0\" ]]; then\n")
	buffer.WriteString("        printf -v ts '%(%s)T' -1\n")
	buffer.WriteString("        echo \"_ #$ts\" >> " + s.Name + ".log\n")
	buffer.WriteString("      " + fmt.Sprintf("    sleep %ds\n", int64(s.Rate.Seconds())))
	buffer.WriteString("      continue\n")
	buffer.WriteString("    fi\n")
	if s.Action.GetScope() == actions.ALL || s.Action.GetScope() == actions.SOME {
		buffer.WriteString("    for (( i=0; i<${#" + inputs + "[@]}; i++ )); do\n")
	} else {
		buffer.WriteString("    for i in $(shuf --input-range=0-$(( ${#" + inputs + "[@]} - 1 )) -n ${sample}); do #takes a subset of size $sample of the " + inputs + "\n")
	}
	buffer.WriteString("      " + input + "=${" + inputs + "[i]}\n")
	buffer.WriteString("      (if " + s.Action.Print() + " ; then\n")
	buffer.WriteString("        printf -v ts '%(%s)T' -1\n")
	buffer.WriteString("        echo \"" + s.Action.Print() + " #$ts\" >> " + s.Name + ".log\n")      //FIXME: escape in action
	buffer.WriteString("        echo \"${" + input + "} ; ${sample} ; $ts\" >> " + s.Name + ".sig\n") //FIXME: escape in action
	buffer.WriteString("      fi)&\n")
	buffer.WriteString("		done\n")
	buffer.WriteString("  " + fmt.Sprintf("    sleep %ds\n", int64(s.Rate.Seconds())))
	buffer.WriteString("  done\n")
	buffer.WriteString("done\n")

	file, err := os.OpenFile(
		s.Name+".sh",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0777,
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer file.Close()
	file.Write(buffer.Bytes())
}
