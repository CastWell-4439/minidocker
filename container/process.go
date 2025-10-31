package container

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

type Process struct {
	Command string
	Args    []string
	Env     []string
	Dir     string
}

func NewProcess(command string, args []string) *Process {
	return &Process{
		Command: command,
		Args:    args,
		Env:     []string{"PATH=/usr/bin:/bin"},
	}
}

func (p *Process) Start() error {
	cmd := exec.Command(p.Command, p.Args...)
	cmd.Env = p.Env
	cmd.Dir = p.Dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, //独立进程组
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("fail to start process,%v", err)
	}
	return cmd.Wait()
}
