package signal

import (
	"bytes"
	"container/list"
	"fmt"
	"os"
)

type Manager struct {
	Name    string
	monkeys *list.List
}

func NewManager(name string) *Manager {
	var manager = Manager{
		Name:    name,
		monkeys: list.New(),
	}
	return &manager
}

func (m *Manager) AddMonkey(monkey *ChaosMonkey) {
	m.monkeys.PushBack(monkey)
}

func (m *Manager) GenerateScripts() {
	var buffer bytes.Buffer
	buffer.WriteString("#! /bin/bash\n")
	buffer.WriteString("trap \"kill 0\" SIGINT SIGTERM\n\n")

	for e := m.monkeys.Front(); e != nil; e = e.Next() {
		var monkey *ChaosMonkey = e.Value.(*ChaosMonkey)
		monkey.GenerateScript()
		buffer.WriteString("(./" + monkey.Name + ".sh) &\n")
	}

	buffer.WriteString("\nwait\n")

	file, err := os.OpenFile(
		m.Name+".sh",
		os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
		0777,
	)
	if err != nil {
		fmt.Println("Error:", err)
	}
	defer file.Close()
	file.Write(buffer.Bytes())
}
