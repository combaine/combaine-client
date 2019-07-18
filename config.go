package main

import (
	"time"

	"github.com/BurntSushi/toml"
)

var lastUpdated = time.Now()

// Tasks is configuration for actions on this client
type Tasks map[string]task
type task struct {
	cmd      string
	interval time.Duration
}

func loadTasks(config string) (Tasks, error) {
	var tasks Tasks
	if _, err := toml.DecodeFile(config, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}
