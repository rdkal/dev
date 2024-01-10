package main

import (
	"context"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Executor struct {
	KillDelay time.Duration
	cmd       string
	args      []string

	terminate func()
	// done is closed if cmd is not running
	done chan struct{}
	c    *exec.Cmd
}

func NewExecutor(cmd []string) *Executor {
	done := make(chan struct{})
	close(done)
	return &Executor{
		cmd:       cmd[0],
		args:      cmd[1:],
		KillDelay: 5 * time.Second,
		terminate: func() {},
		done:      done,
	}
}

func (e *Executor) isDone() bool {
	select {
	case <-e.done:
		return true
	default:
		return false
	}
}

func (e *Executor) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-e.done:
		return nil
	}
}

func (e *Executor) Restart(ctx context.Context) error {
	e.terminate()
	err := e.Wait(ctx)
	if err != nil {
		return err
	}
	return e.Start(ctx)
}

func (e *Executor) Start(ctx context.Context) error {
	if !e.isDone() {
		return nil
	}

	cmd := exec.Command(e.cmd, e.args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return err
	}
	pgid, err := syscall.Getpgid(cmd.Process.Pid)
	if err != nil {
		return err
	}
	done := make(chan struct{})

	ctx, terminate := context.WithCancel(ctx)
	go func() {
		<-ctx.Done()
		syscall.Kill(-pgid, syscall.SIGINT)
		select {
		case <-done:
		case <-time.After(e.KillDelay):
			syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}()

	go func() {
		cmd.Wait()
		close(done)
	}()

	e.c = cmd
	e.done = done
	e.terminate = terminate
	return nil
}
