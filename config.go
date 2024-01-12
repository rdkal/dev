package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/BurntSushi/toml"
)

const ConfigName = ".dev.toml"

type ConfigExec struct {
	Command []string `toml:"cmd"`
}

type ConfigWatcher struct {
	ExcludeGlobs []string `toml:"exclude"`
}

type ConfigServer struct {
	DevServerPort int    `toml:"port"`
	FowardToURL   string `toml:"forward_to_url"`
}

type Config struct {
	ConfigExec    `toml:"exec"`
	ConfigWatcher `toml:"watcher"`
	ConfigServer  `toml:"server"`
}

func DefaultConfig() *Config {
	return &Config{
		ConfigExec: ConfigExec{
			Command: []string{"go", "run", "."},
		},
		ConfigWatcher: ConfigWatcher{
			ExcludeGlobs: []string{"*_test.go", "*_templ.go", ".git", ConfigName},
		},
		ConfigServer: ConfigServer{
			DevServerPort: 8081,
			FowardToURL:   "http://localhost:8080",
		},
	}
}

func GetConfig() (*Config, error) {
	cfg := DefaultConfig()
	_, err := toml.DecodeFile(ConfigName, cfg)
	if errors.Is(err, fs.ErrNotExist) {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func InitConifg() error {
	_, err := os.Stat(ConfigName)
	if err == nil {
		return fmt.Errorf("%s already exists", ConfigName)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	f, err := os.Create(ConfigName)
	if err != nil {
		return err
	}
	defer f.Close()

	cfg := DefaultConfig()
	err = toml.NewEncoder(f).Encode(cfg)
	if err != nil {
		return err
	}

	return f.Close()
}
