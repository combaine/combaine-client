package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

var defaultTimeout = time.Minute

type clientSettings struct {
	Gzip bool `yaml:",omitempty"`
	Port int
}

type clientConfig struct {
	Settings clientSettings
	Tasks    []*task
}
type tasksStore map[string]*task

type configLoader struct {
	m          sync.RWMutex
	configFile string
	settings   clientSettings
	store      tasksStore
	mtime      time.Time
	stop       chan bool
	log        echo.Logger
}

func newConfigLoader(configFile string, log echo.Logger) (*configLoader, error) {
	var l = &configLoader{
		configFile: configFile,
		store:      make(tasksStore),
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

func (l *configLoader) load() (*clientSettings, tasksStore, error) {
	bytes, err := ioutil.ReadFile(l.configFile)
	if err != nil {
		return nil, nil, err
	}
	var c clientConfig
	c.Settings.Gzip = true
	if err := yaml.Unmarshal(bytes, &c); err != nil {
		return nil, nil, err
	}
	store := make(tasksStore)
	for idx, t := range c.Tasks {
		if t.Name == "" {
			return nil, nil, errors.Errorf("Task idx=%d has no name", idx)
		}
		if t.Interval != "" {
			d, err := time.ParseDuration(t.Interval)
			if err != nil {
				return nil, nil, errors.Wrap(err, t.Name+".Interval")
			}
			if d < 100*time.Millisecond {
				return nil, nil, errors.New(t.Name + ".Interval too small (min 100ms): " + t.Interval)
			}
			t.intervalDuration = d
		}
		if t.Splice != "" {
			d, err := time.ParseDuration(t.Splice)
			if err != nil {
				return nil, nil, errors.Wrap(err, t.Name+".Splice")
			}
			if d < 10*time.Millisecond {
				return nil, nil, errors.New(t.Name + ".Splice too small (min 10ms): " + t.Splice)
			}
			t.spliceDuration = d
		}
		if t.Timeout != "" {
			d, err := time.ParseDuration(t.Timeout)
			if err != nil {
				return nil, nil, errors.Wrap(err, t.Name+".Timeout")
			}
			if t.Interval != "" && d > t.intervalDuration {
				return nil, nil, errors.New(t.Name + ".Timeout > " + t.Name + ".Interval")
			}
			t.timeoutDuration = d
		} else {
			if t.Interval != "" {
				t.timeoutDuration = t.intervalDuration
			} else {
				t.timeoutDuration = defaultTimeout
			}
		}
		store[t.Name] = t
	}
	return &c.Settings, store, nil
}

func (l *configLoader) reload() error {
	s, store, err := l.load()
	if err != nil {
		return err
	}
	l.m.Lock()
	l.settings = *s
	l.store = store
	l.mtime = time.Now()
	l.m.Unlock()
	return nil
}

func (l *configLoader) tasksList() string {
	var buf bytes.Buffer
	l.m.RLock()
	for k := range l.store {
		buf.WriteString(k + "\n")
	}
	l.m.RUnlock()
	return buf.String()
}

func (l *configLoader) lookupTask(taskName string) (*task, bool) {
	l.m.RLock()
	t, ok := l.store[taskName]
	l.m.RUnlock()
	return t, ok
}
