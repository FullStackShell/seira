package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".seirarc.json"

type Config struct {
	Entrypoint string            `json:"entrypoint"`
	Shebang    string            `json:"shebang"`
	Env        map[string]string `json:"env"`
	Include    []string          `json:"include"`
	Exclude    []string          `json:"exclude"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Shebang: "/bin/sh",
		Env:     map[string]string{},
		Include: []string{},
		Exclude: []string{},
	}
}

// Load reads .seirarc.json by walking upward from dir.
// If no config file is found, returns Default() with no error.
func Load(dir string) (*Config, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	for {
		p := filepath.Join(absDir, configFileName)
		data, err := os.ReadFile(p)
		if err == nil {
			return parse(data)
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		absDir = parent
	}

	return Default(), nil
}

func parse(data []byte) (*Config, error) {
	cfg := Default()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.Env == nil {
		cfg.Env = map[string]string{}
	}
	if cfg.Include == nil {
		cfg.Include = []string{}
	}
	if cfg.Exclude == nil {
		cfg.Exclude = []string{}
	}
	return cfg, nil
}
