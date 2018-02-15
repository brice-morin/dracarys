package actions

import (
	"fmt"
)

/*Limits quota allocated to a container*/
type RedisBench struct {
	Action
	Test string
}

func (a *RedisBench) Default() IAction {
	a.SetScope(SOME)
	a.SetType(CONTAINER)
	return a
}

func (a *RedisBench) Print() string {
	cmd := "docker run -it --rm --link MyRedisContainer:redis clue/redis-benchmark " +
		"-t %s " +
		"-c ${client} " +
		"-d ${size} " +
		"-r ${keyset} " +
		"-n 10000 " +
		"-P ${pipeline} " +
		"--csv | while IFS= read -r line; do printf '%%s,%%s\\n' \"$(date '+%%s')\" \"$line\"; done > bench.log"
	return fmt.Sprintf(cmd, a.Test)
}
