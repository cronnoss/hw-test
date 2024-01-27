package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "./configs/scheduler_config.toml", "Path to configuration file")
	flag.Parse()

	if flag.Arg(0) == "version" {
		app.PrintVersion()
		return
	}
}

type Config struct {
	app.SchedulerConf
}

func NewConfig() Config {
	var config Config
	if err := config.LoadConfigFile(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Can't load config file:%v error: %v\n", configFile, err)
		os.Exit(1)
	}
	fmt.Println("Config:", config)
	return config
}

func (c *Config) LoadConfigFile(filename string) error {
	filedata, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return toml.Unmarshal(filedata, c)
}
