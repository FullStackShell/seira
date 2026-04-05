package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const configFileName = ".seirarc.json"

type Config struct {
	Entrypoint   string            `json:"entrypoint"`
	Shebang      string            `json:"shebang"`
	Mode         string            `json:"mode"`      // "tarball" or "concat", default "tarball"
	Type         string            `json:"type"`      // "executable" (default) or "library"
	Env          map[string]string `json:"env"`
	Include      []string          `json:"include"`
	Exclude      []string          `json:"exclude"`
	Dependencies map[string]string `json:"dependencies"` // bpkg-style: "user/name": "version"
	Exports      []string          `json:"exports"`      // library mode: exported function names (empty = all)
	DepsDir      string            `json:"deps_dir"`     // deps directory (default: "deps")
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Shebang:      "/bin/sh",
		Env:          map[string]string{},
		Include:      []string{},
		Exclude:      []string{},
		Dependencies: map[string]string{},
		Exports:      []string{},
		DepsDir:      "deps",
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
	if cfg.Dependencies == nil {
		cfg.Dependencies = map[string]string{}
	}
	if cfg.Exports == nil {
		cfg.Exports = []string{}
	}
	if cfg.DepsDir == "" {
		cfg.DepsDir = "deps"
	}
	return cfg, nil
}
