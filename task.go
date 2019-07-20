package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

type task struct {
	Name             string
	Cmd              string
	Timeout          string
	timeoutDuration  time.Duration
	Interval         string
	intervalDuration time.Duration
	Splice           string
	spliceDuration   time.Duration
}

func (t *task) getOutput(rid string, c echo.Context) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), t.timeoutDuration)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/bin/bash")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdin = strings.NewReader(t.Cmd)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	go func() {

		<-ctx.Done()
		err := ctx.Err()
		switch err {
		case context.DeadlineExceeded:
			c.Logger().Warnj(log.JSON{"id": rid, "msg": "Kill task", "reason": err.Error()})
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
	}()
	c.Logger().Infoj(log.JSON{"id": rid, "msg": "Run", "task": fmt.Sprintf("%#v", t)})
	if err := cmd.Run(); err != nil {
		c.Logger().Errorj(log.JSON{
			"id":     rid,
			"stderr": stderr.String(),
			"stdout": stdout.String(),
			"error":  err.Error(),
		})
		return "", err
	}
	return stdout.String(), nil
}
