package main

import (
	"time"

	"github.com/labstack/echo"
)

type task struct {
	Name             string
	Cmd              string
	Interval         string
	intervalDuration time.Duration
	Splice           string
	spliceDuration   time.Duration
}

func (t *task) getOutput(l echo.Logger) string {
	return "Invoke task " + t.Name + ": " + t.Cmd
}
