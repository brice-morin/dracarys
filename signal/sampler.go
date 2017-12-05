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
	samples *ring.Ring
	Signal
	Rate   time.Duration
	Action actions.IAction
	Ticker *time.Ticker
}

func NewChaosMonkey(sig *Signal, r time.Duration, a actions.IAction /*docker.DockerAction*/) *ChaosMonkey {
	a.SetTick(r)
	var sampler ChaosMonkey = ChaosMonkey{
		Signal: *sig,
		Rate:   r,
		Action: a,
	}
	return &sampler
}

func (s *ChaosMonkey) Init() ChaosMonkey {
	size := int(s.Signal.GetPeriod().Seconds()) / int(s.Rate.Seconds())
	s.samples = ring.New(size)
	for i := 0; i < int(s.Signal.GetPeriod().Seconds()); i = i + int(s.Rate.Seconds()) {
		s.samples.Value = s.Signal.Sample(int64(i))
		fmt.Println("Sampling signal[", i, "]", s.samples.Value)
		s.samples = s.samples.Next()
	}
	s.Signal = nil
	return *s
}

func (s *ChaosMonkey) GenerateScript(path string) {
	var buffer bytes.Buffer
	buffer.WriteString("#! /bin/bash\n")
	buffer.WriteString("alive=true\n")
	buffer.WriteString("_term(){\n")
	buffer.WriteString("alive=false\n")
	buffer.WriteString("  echo \"Waiting " + fmt.Sprintf("%ds", int64(s.Rate.Seconds())) + "for processes to terminate...\"\n")
	buffer.WriteString(fmt.Sprintf("  sleep %ds\n", int64(s.Rate.Seconds())))
	buffer.WriteString("  for job in `jobs -p`\n")
	buffer.WriteString("  do\n")
	buffer.WriteString("    echo $job\n")
	buffer.WriteString("    wait $job\n")
	buffer.WriteString("  done\n")
	//TODO: add cleaning operations (restablishing containers in a decent state when we will the script)
	//NOTE: though ideally, operations will leave the container in its initial state after they finish, so stopping the while loop and waiting for processes should be OK...
	buffer.WriteString("  echo \"The End.\"\n")
	buffer.WriteString("  exit\n")
	buffer.WriteString("}\n\n")

	buffer.WriteString("trap _term INT\n\n")

	buffer.WriteString("declare -a samples=(")
	for i := 0; i < int(s.Signal.GetPeriod().Seconds()); i = i + int(s.Rate.Seconds()) {
		if i > 0 {
			buffer.WriteString(" ")
		}
		buffer.WriteString(fmt.Sprintf("%d", s.Signal.Sample(int64(i))))
	}
	buffer.WriteString(")\n")
	buffer.WriteString("while $alive; do\n")
	buffer.WriteString("  for sample in \"${samples[@]}\"; do\n")
	buffer.WriteString("    echo $sample\n")
	buffer.WriteString("    for container in `docker ps -q --format \"{{.Names}}\" --filter \"label=dracarys=true\"\"$@\"`; do\n")
	buffer.WriteString("      (if " + s.Action.Print() + " ; then\n")
	buffer.WriteString("        printf -v ts '%(%s)T' -1\n")
	buffer.WriteString("        echo \"" + s.Action.Print() + " #$ts\" >> dracarys.log\n") //FIXME: escape in action
	buffer.WriteString("      fi)&\n")
	buffer.WriteString("		done\n")
	buffer.WriteString(fmt.Sprintf("    sleep %ds\n", int64(s.Rate.Seconds())))
	buffer.WriteString("  done\n")
	buffer.WriteString("done\n")

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0777,
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer file.Close()
	file.Write(buffer.Bytes())
}

func (s *ChaosMonkey) Start(c chan Sample, done *chan bool) {
	var now int64 = time.Now().Unix()
	s.Ticker = time.NewTicker(s.Rate)
	for t := range s.Ticker.C {
		ts := t.Unix() - now
		v := s.samples.Value.(int64)
		s.samples = s.samples.Next()
		fmt.Println("Tick[", ts, "]:", v)
		go s.Action.Do(v)
		c <- Sample{
			ts,
			v,
		}
	}
	*done <- true
}

func (s *ChaosMonkey) Stop() {
	s.Ticker.Stop()
}
