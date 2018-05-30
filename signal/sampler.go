package signal

import (
	"bytes"
	"container/ring"
	"fmt"
	"math"
	"os"
	"strings"
	"time"

	"../actions"
)

type Sample struct {
	t int64
	v int64
}

type ChaosMonkey struct {
	Name    string
	samples *ring.Ring
	Signals []*Signal
	Shift   time.Duration
	Rate    time.Duration
	Action  actions.IAction
	Ticker  *time.Ticker
}

func NewChaosMonkey(name string, sig []*Signal, shift time.Duration, r time.Duration, a actions.IAction) *ChaosMonkey {
	a.SetTick(r)
	var sampler = ChaosMonkey{
		Name:    name,
		Signals: sig,
		Rate:    r,
		Action:  a,
		Shift:   shift,
	}
	return &sampler
}

func printTargets(a actions.IAction) string {
	var buffer bytes.Buffer
	for i, target := range a.GetTargets() {
		if i > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString("\"" + target + "\"")
	}
	return buffer.String()
}

func generateHeader(buffer *bytes.Buffer, rate time.Duration) {
	buffer.WriteString("#! /bin/bash\n")
	buffer.WriteString("alive=true\n")
	buffer.WriteString("function _term {\n")
	buffer.WriteString("  alive=false\n")
	buffer.WriteString("  echo \"Waiting " + fmt.Sprintf("%ds", int64(rate.Seconds())) + "for processes to terminate... Ctrl+C again to quit now.\"\n")
	buffer.WriteString(fmt.Sprintf("  sleep %ds\n", int64(rate.Seconds())))
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
}

func generateSamples(buffer *bytes.Buffer, s *ChaosMonkey) {
	for _, signal := range s.Signals {
		sig := *signal
		name := sig.GetVariable().Name
		name = strings.Replace(name, "-", "_", -1)
		buffer.WriteString("declare -a all_" + name + "=(")
		values := []string{}
		for i := int64(s.Shift.Seconds()); i < int64(s.Shift.Seconds()+sig.GetPeriod().Seconds()); i = i + int64(s.Rate.Seconds()) {
			v := sig.Sample(i)
			if math.Floor(v)-v == 0 {
				values = append(values, fmt.Sprintf("%d", int64(v)))
			} else {
				values = append(values, fmt.Sprintf("%.2f", v))
			}
		}
		buffer.WriteString(strings.Join(values, " "))
		buffer.WriteString(")\n")
	}
}

func generateTargets(buffer *bytes.Buffer, a actions.IAction) {
	if a.GetScope() == actions.SOME {
		buffer.WriteString("declare -a targets=(" + printTargets(a) + ")\n")
	}
}

func generateFindTarget(buffer *bytes.Buffer, s *ChaosMonkey) {
	if s.Action.GetScope() != actions.NEW {
		switch s.Action.GetType() {
		case actions.CONTAINER:
			if s.Action.GetScope() != actions.SOME {
				buffer.WriteString("    mapfile -t targets < <(docker ps -q --format \"{{.Names}}\" --filter label=eu.stamp.dracarys=true)\n") //TODO --filter "label=dracarys=true"
			}
		case actions.NETWORK:
			if s.Action.GetScope() != actions.SOME {
				buffer.WriteString("    mapfile -t targets < <(docker network ls -q --format \"{{.Name}}\")\n") //TODO --filter "label=dracarys=true"
			}
		case actions.SERVICE:
			if s.Action.GetScope() != actions.SOME {
				buffer.WriteString("    mapfile -t targets < <(docker service ls -q --format \"{{.Name}}\" --filter label=eu.stamp.dracarys=true)\n") //TODO --filter "label=dracarys=true"
			}
		}
		buffer.WriteString("    if [[ ${#targets[@]} -eq \"0\" ]]; then\n")
		buffer.WriteString("      printf -v ts '%(%s)T' -1\n")
		buffer.WriteString("      echo \"$ts,_,no target\" >> " + s.Name + ".log\n")
		buffer.WriteString("      " + fmt.Sprintf("sleep %ds\n", int64(s.Rate.Seconds())))
		buffer.WriteString("      continue\n")
		buffer.WriteString("    fi\n")
	}
}

func generateApplyToTargets(buffer *bytes.Buffer, s *ChaosMonkey) {
	if s.Action.GetScope() != actions.NEW {
		if s.Action.GetScope() == actions.ALL || s.Action.GetScope() == actions.SOME {
			buffer.WriteString("    for (( i=0; i<${#targets[@]}; i++ )); do\n")
		} else {
			buffer.WriteString("    for i in $(shuf --input-range=0-$(( ${#targets[@]} - 1 )) -n ${sample}); do #takes a subset of size $sample of the targets\n")
		}
		buffer.WriteString("      target=${targets[i]}\n")
	}

	values := ""
	for _, v := range s.Action.GetVariables() {
		name := strings.Replace(v.Name, "-", "_", -1)
		buffer.WriteString("      " + name + "=${all_" + name + "[$index%${#all_" + name + "[@]}]}\n")
		values += "${" + name + "},"
	}
	buffer.WriteString("      error=`" + s.Action.Print() + " 2>&1 >> " + s.Action.GetOutput() + "`\n")
	buffer.WriteString("      printf -v ts '%(%s)T' -1\n")
	buffer.WriteString("      if [ $? -eq 0 ]; then\n")
	buffer.WriteString("        echo \"$ts," + s.Action.Print() + ",${error}\" >> " + s.Name + ".log\n")
	buffer.WriteString("        echo \"$ts," + values + "${target}\" >> " + s.Name + ".sig\n")
	buffer.WriteString("      else\n")
	buffer.WriteString("        echo \"$ts," + s.Action.Print() + ",${error}\" >> " + s.Name + ".log\n")
	buffer.WriteString("      fi\n")

	if s.Action.GetScope() != actions.NEW {
		buffer.WriteString("		done\n")
	}
}

func (s *ChaosMonkey) GenerateScript() {
	var buffer bytes.Buffer
	generateHeader(&buffer, s.Rate)
	buffer.WriteString(s.Action.PrintHelper() + "\n")
	generateSamples(&buffer, s)
	generateTargets(&buffer, s.Action)

	buffer.WriteString("index=0\n")
	buffer.WriteString("while $alive; do\n")
	buffer.WriteString("start=`date +\"%s\"`\n")

	generateFindTarget(&buffer, s)
	generateApplyToTargets(&buffer, s)

	buffer.WriteString("    stop=`date +\"%s\"`\n")
	buffer.WriteString("    duration=$(( $stop-$start ))\n")
	buffer.WriteString("    stop=`date +\"%s\"`\n")
	buffer.WriteString("    " + fmt.Sprintf("sleep_time=$(( %d-$duration ))\n", int64(s.Rate.Seconds())))
	buffer.WriteString("    while [ $sleep_time -lt \"0\" ]\n")
	buffer.WriteString("    do\n")
	buffer.WriteString("      " + fmt.Sprintf("duration=$(( $duration+%d ))\n", int64(s.Rate.Seconds())))
	buffer.WriteString("    done\n")
	buffer.WriteString("    sleep ${sleep_time}s\n")
	buffer.WriteString("    ((index++))\n")
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
