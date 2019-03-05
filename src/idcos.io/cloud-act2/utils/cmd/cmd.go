//  Copyright (c) Cloud J Tech, Inc. All rights reserved.
//  Licensed under the GPLv3 License. See License.txt in the project root for license information.
package cmd

import (
	"bytes"
	"errors"
	"os/exec"
	"time"

	"github.com/hashicorp/go-hclog"
	"idcos.io/cloud-act2/log"
	"io"
)

var (
	//ErrTimeout timeout error
	ErrTimeout = errors.New("timeout")
)

func getLogger() hclog.Logger {
	return log.L().Named("utils.cmd")
}

// Result execute result
type Result struct {
	Executor *Executor
	Stdout   *bytes.Buffer
	Stderr   *bytes.Buffer
	ExitCode int
	Error    error
}

//Executor executor
type Executor struct {
	Cmd   string
	Args  []string
	Stdin *bytes.Buffer
}

//NewExecutor build new executor
func NewExecutor(cmd string, args ...string) *Executor {
	return &Executor{
		Cmd:  cmd,
		Args: args,
	}
}

// InvokeWithTimeout  with timeout
// @param: timeout, second
func (e *Executor) InvokeWithTimeout(timeout time.Duration) *Result {
	logger := getLogger()

	cmd := exec.Command(e.Cmd, e.Args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if e.Stdin != nil {
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			return nil
		}

		go func() {
			defer stdinPipe.Close()
			io.Copy(stdinPipe, e.Stdin)
		}()
	}

	cmd.Start()

	r := Result{
		Executor: e,
		Stdout:   &stdout,
		Stderr:   &stderr,
	}
	done := make(chan struct{})
	go func() {
		err := cmd.Wait()
		if err != nil {
			r.Error = err
			logger.Info("run command error ", hclog.Fmt("%s", err))
		}
		close(done)
	}()

	// for timeout
	ticker := time.NewTicker(timeout * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			logger.Debug("run command done")
			return &r
		default:
			// none forbidden
		}

		select {
		case <-ticker.C:
			r.Error = ErrTimeout
			return &r
		default:
			// none forbidden
		}
	}

	// return &r
}

// Invoke invoke
func (e *Executor) Invoke() *Result {
	cmd := exec.Command(e.Cmd, e.Args...)

	var stdout, stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if e.Stdin != nil {
		stdinPipe, err := cmd.StdinPipe()
		if err != nil {
			return nil
		}

		go func() {
			defer stdinPipe.Close()
			io.Copy(stdinPipe, e.Stdin)
		}()
	}
	err := cmd.Run()

	r := Result{
		Executor: e,
		Stdout:   &stdout,
		Stderr:   &stderr,
		Error:    err,
	}
	return &r
}
