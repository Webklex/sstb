package app

import (
	"../utils/config"
	"../utils/log"
	"./notifier"
	"flag"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"os"
	"path"
	"path/filepath"
	"time"
)

type App struct {
	*Config

	jobs []*Job
}

func DefaultConfig() *Config {
	dir, _ := os.Getwd()

	c := &Config{
		Config:   config.DefaultConfig(),
		Timezone: "UTC",
		JobDir:   path.Join(dir, "config", "jobs"),
		Build:    Build{},
		Provider: make([]*Provider, 0),
	}
	c.File = path.Join(dir, "config", "app.json")
	c.Config.SetContext(c)

	return c
}

func NewConfigFromFile(configFile string) *Config {
	c := DefaultConfig()

	c.Load(configFile)
	c.File = configFile

	return c
}

// AddFlags adds configuration flags to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	defer envconfig.Process("sstb", c)

	fs.StringVar(&c.Timezone, "timezone", c.Timezone, "Application time zone")
	fs.StringVar(&c.JobDir, "job-dir", c.JobDir, "Folder containing all job configuration files")
	fs.StringVar(&c.File, "config", c.File, "Application config file")
}

func NewApp(c *Config) *App {
	a := &App{
		Config: c,
		jobs:   make([]*Job, 0),
	}

	for _, n := range a.Notifier {
		n.Init()
	}

	a.loadJobs()

	return a
}

func (a *App) loadJobs() {
	a.jobs = make([]*Job, 0)

	_ = filepath.Walk(a.JobDir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".json" {

			j := NewJobFromFile(path)
			p := a.getProvider(j.ProviderId)
			j.Notifier = a.getNotifiers(j.NotifierIds)

			if p == nil {
				log.Error(fmt.Sprintf("Unkown provider: %s", j.ProviderId))
			} else {
				j.setProvider(p)
				a.jobs = append(a.jobs, j)
			}
		}
		return nil
	})

	log.Success(fmt.Sprintf("Loaded %d jobs", len(a.jobs)))
}

func (a *App) getProvider(key string) *Provider {
	for _, p := range a.Provider {
		if p.Name == key {
			return p
		}
	}
	return nil
}

func (a *App) getNotifiers(keys []string) []*notifier.Notifier {
	ns := make([]*notifier.Notifier, 0)
	for _, key := range keys {
		for _, n := range a.Notifier {
			if n.Name == key {
				ns = append(ns, n)
			}
		}
	}
	return ns
}

func (a *App) Start() {
	for _, j := range a.jobs {
		if j.Enabled == false {
			continue
		}

		go j.Start()
	}

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:

			for _, j := range a.jobs {
				if j.Enabled == false {
					continue
				}

				go j.Tick(t)
			}
		}
	}

}
