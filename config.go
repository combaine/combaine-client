package main

import (
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

var lastUpdated = time.Now()

type tasksStore map[string]*task

type configLoader struct {
	m          sync.RWMutex
	store      tasksStore
	configFile string
	stop       chan bool
	log        echo.Logger
	mtime      time.Time
}

func newConfigLoader(configFile string, log echo.Logger) (*configLoader, error) {
	var l = &configLoader{
		store:      make(tasksStore),
		configFile: configFile,
		stop:       make(chan bool),
		log:        log,
	}

	err := l.reload()
	go l.periodicReload()
	return l, err
}

func (l *configLoader) periodicReload() {
	t := time.NewTicker(1 * time.Minute)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			stat, err := os.Stat(l.configFile)
			if err != nil {
				l.log.Error("Failed to get stat for %s: %v", l.configFile, err)
			}
			if stat.ModTime().After(l.mtime) {
				if err := l.reload(); err != nil {
					l.log.Error("Failed to reload %s: %v", l.configFile, err)
				}
			}
		case <-l.stop:
			return
		}
	}
}

func (l *configLoader) load() (tasksStore, error) {
	var store tasksStore
	if _, err := toml.DecodeFile(l.configFile, &store); err != nil {
		return nil, err
	}
	// update task.name
	for k, t := range store {
		if t.Name == "" {
			t.Name = k
		}
		if t.Interval != "" {
			d, err := time.ParseDuration(t.Interval)
			if err != nil {
				return nil, errors.Wrap(err, k+".Interval")
			}
			t.intervalDuration = d
		}
		if t.Splice != "" {
			d, err := time.ParseDuration(t.Splice)
			if err != nil {
				return nil, errors.Wrap(err, k+".Splice")
			}
			t.spliceDuration = d
		}
	}
	return store, nil
}

func (l *configLoader) reload() error {
	store, err := l.load()
	if err != nil {
		return err
	}
	l.m.Lock()
	l.store = store
	l.mtime = time.Now()
	l.m.Unlock()
	return nil
}

func (l *configLoader) lookup(taskName string) (*task, bool) {
	l.m.RLock()
	t, ok := l.store[taskName]
	l.m.RUnlock()
	return t, ok
}
