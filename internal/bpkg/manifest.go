package bpkg

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Manifest represents a bpkg.json package manifest.
type Manifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Scripts      []string          `json:"scripts"`
	Files        []string          `json:"files"`
	Dependencies map[string]string `json:"dependencies"`
	Install      string            `json:"install"`
	Global       string            `json:"global"`
}

// LoadManifest reads a bpkg.json (or package.json fallback) from dir.
func LoadManifest(dir string) (*Manifest, error) {
	// Try bpkg.json first, then package.json as fallback
	for _, name := range []string{"bpkg.json", "package.json"} {
		p := filepath.Join(dir, name)
		data, err := os.ReadFile(p)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		return &m, nil
	}
	return nil, os.ErrNotExist
}
