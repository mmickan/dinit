package main

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// Primary holds which pid is considered the primary process. If that
// dies, the whole container should be killed.
type Primary struct {
	sync.RWMutex
	first map[int]bool
	all   bool
}

func NewPrimary() *Primary {
	return &Primary{first: make(map[int]bool)}
}

func (p *Primary) Set(pid int) {
	p.Lock()
	defer p.Unlock()
	p.first[pid] = true
}

func (p *Primary) Primary(pid int) bool {
	p.RLock()
	defer p.RUnlock()
	_, ok := p.first[pid]
	return ok
}

func (p *Primary) All() bool {
	p.RLock()
	defer p.RUnlock()
	return p.all
}

func (p *Primary) SetAll(all bool) {
	p.Lock()
	defer p.Unlock()
	p.all = all
}

// Procs holds the processes that we run.
type Procs struct {
	sync.RWMutex
	pids map[int]*exec.Cmd
}

func NewProcs() *Procs {
	return &Procs{pids: make(map[int]*exec.Cmd)}
}

func (c *Procs) Insert(cmd *exec.Cmd) {
	c.Lock()
	defer c.Unlock()
	c.pids[cmd.Process.Pid] = cmd
}

func (c *Procs) Remove(cmd *exec.Cmd) {
	c.Lock()
	defer c.Unlock()
	delete(c.pids, cmd.Process.Pid)
}

// Signal sends sig to all processes in Procs.
func (c *Procs) Signal(sig os.Signal) {
	c.RLock()
	defer c.RUnlock()
	for pid, cmd := range c.pids {
		if test.Test() {
			lg.Printf("signal %d sent to pid %d", sig, testPid)
		} else {
			lg.Printf("signal %d sent to pid %d", sig, pid)
		}
		cmd.Process.Signal(sig)
	}
}

// Cleanup will send signal sig to the processes and after a short time send a SIGKKILL.
func (c *Procs) Cleanup(sig os.Signal) {
	if prestop != "" {
		prestopcmd := command(prestop)
		if err := prestopcmd.Start(); err != nil {
			lg.Fatalf("prestop command failed to start: %v", err)
		}
		pid := prestopcmd.Process.Pid
		if test.Test() {
			pid = testPid
		}
		lg.Printf("pid %d started: %v", pid, prestopcmd.Args)
		time.Sleep(prestoptimer)
	}

	procs.Signal(sig)

	time.Sleep(2 * time.Second)

	if procs.Len() > 0 {
		lg.Printf("%d processes still alive after SIGINT/SIGTERM", procs.Len())
		time.Sleep(timeout)
	}
	procs.Signal(syscall.SIGKILL)
}

// Len returns the number of processs in Procs.
func (c *Procs) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.pids)
}
