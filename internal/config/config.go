package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var configFileNames = []string{"seirarc.json", ".seirarc.json"}

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
	TreeShake    bool              `json:"treeshake"`    // enable tree shaking
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
	cfg, _, err := LoadWithDir(dir)
	return cfg, err
}

// tryLoadFrom tries to load a config file from the given directory.
// It checks both seirarc.json and .seirarc.json (in that order).
// Returns the parsed config, true if found, or an error.
func tryLoadFrom(dir string) (*Config, bool, error) {
	for _, name := range configFileNames {
		p := filepath.Join(dir, name)
		data, err := os.ReadFile(p)
		if err == nil {
			cfg, err := parse(data)
			return cfg, true, err
		}
		if !os.IsNotExist(err) {
			return nil, false, err
		}
	}
	return nil, false, nil
}

// LoadWithDir reads seirarc.json or .seirarc.json by walking upward from dir to the root.
// Returns the parsed config and the directory where the config file was found.
// If no config file is found, returns Default(), empty string, and no error.
func LoadWithDir(dir string) (*Config, string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}

	for {
		cfg, found, err := tryLoadFrom(absDir)
		if err != nil {
			return nil, "", err
		}
		if found {
			return cfg, absDir, nil
		}
		parent := filepath.Dir(absDir)
		if parent == absDir {
			break
		}
		absDir = parent
	}

	return Default(), "", nil
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
