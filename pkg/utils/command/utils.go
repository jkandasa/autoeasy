package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/go-cmd/cmd"
	"go.uber.org/zap"
)

// command exit types
const (
	ExitTypeNormal  = "normal"
	ExitTypeTimeout = "timeout"
	ExitTypeStop    = "stop"
)

// Command deails
type Command struct {
	Name                 string
	Command              string
	Args                 []string
	Env                  []string
	Timeout              time.Duration
	StatusUpdateDuration time.Duration
	StatusFn             func(cmd.Status)
	ExitFn               func(string, cmd.Status)
	isRunning            bool
	mutex                sync.RWMutex
	cmd                  *cmd.Cmd
	stopCh               chan bool
}

// Status returns the current status of the command
func (c *Command) Status() cmd.Status {
	return c.cmd.Status()
}

// IsRunning returns the current status of the command
func (c *Command) IsRunning() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.isRunning
}

func (c *Command) setRunning(status bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.isRunning = status
}

// Start triggers the command
func (c *Command) Start() error {
	if c.IsRunning() {
		return errors.New("start already triggered")
	}
	go c.startFn()
	return nil
}

// StartAndWait triggers the command and wait till it completes
func (c *Command) StartAndWait() error {
	if c.IsRunning() {
		return errors.New("start already triggered")
	}
	c.startFn()
	return nil
}

// Stop terminates the command
func (c *Command) Stop() error {
	if !c.isRunning {
		return errors.New("not yet started")
	}
	c.stopCh <- true
	return nil
}

func (c *Command) startFn() {
	c.setRunning(true) // set as running

	// update status func, if not set
	if c.StatusFn == nil {
		c.StatusFn = PrintStatus
	}

	c.cmd = cmd.NewCmd(c.Command, c.Args...)
	if len(c.Env) > 0 {
		c.cmd.Env = c.Env
	}
	statusChan := c.cmd.Start()
	zap.L().Debug("command execution started", zap.String("command", c.Command))

	doneCh := make(chan bool)
	defer close(doneCh)

	// update on exit
	onExitFn := func(exitType string, status cmd.Status) {

		// terminate status update goroutine
		if c.StatusFn != nil {
			doneCh <- true
		}
		if c.ExitFn != nil {
			c.ExitFn(exitType, status)
		}
		// update it the running status
		c.setRunning(false)
	}

	// status update function
	if c.StatusFn != nil {
		go func(done <-chan bool) {
			statusTicker := time.NewTicker(c.StatusUpdateDuration)
			defer statusTicker.Stop()
			for {
				select {
				case <-done:
					return
				case <-statusTicker.C:
					c.StatusFn(c.cmd.Status())
				}
			}
		}(doneCh)
	}

	select {
	case <-c.stopCh: // stop triggered
		zap.L().Debug("command execution stop triggered", zap.String("command", c.Command))

		err := c.cmd.Stop()
		if err != nil {
			zap.L().Debug("error on stopping command execution", zap.String("command", c.Command), zap.Error(err))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeStop, st)

	case <-time.After(c.Timeout): // timeout
		zap.L().Debug("command execution reached timeout", zap.String("command", c.Command), zap.String("timeout", c.Timeout.String()))
		err := c.cmd.Stop()
		if err != nil {
			zap.L().Debug("error on stopping command execution", zap.String("command", c.Command), zap.Error(err))
		}
		st := c.cmd.Status()
		onExitFn(ExitTypeTimeout, st)

	case status := <-statusChan: // command execution completed
		zap.L().Debug("command execution completed", zap.String("command", c.Command))
		onExitFn(ExitTypeNormal, status)
	}
}

func PrintStatus(status cmd.Status) {
	zap.L().Debug("command execution status", zap.Int("pid", status.PID), zap.String("command", status.Cmd), zap.Bool("isComplete", status.Complete), zap.Error(status.Error))
}
