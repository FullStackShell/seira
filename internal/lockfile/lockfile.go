package lockfile

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const fileName = ".seira-lock.json"

// LockEntry records the resolved version and commit hash of a dependency.
type LockEntry struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// LockFile records resolved dependency information for reproducible installs.
type LockFile struct {
	Dependencies map[string]LockEntry `json:"dependencies"`
}

// New creates an empty LockFile.
func New() *LockFile {
	return &LockFile{
		Dependencies: map[string]LockEntry{},
	}
}

// Load reads .seira-lock.json from dir. Returns a new empty LockFile if not found.
func Load(dir string) (*LockFile, error) {
	p := filepath.Join(dir, fileName)
	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}

	lf := New()
	if err := json.Unmarshal(data, lf); err != nil {
		return nil, err
	}
	if lf.Dependencies == nil {
		lf.Dependencies = map[string]LockEntry{}
	}
	return lf, nil
}

// Save writes .seira-lock.json to dir.
func (l *LockFile) Save(dir string) error {
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	p := filepath.Join(dir, fileName)
	return os.WriteFile(p, append(data, '\n'), 0644)
}

// Set records a dependency's version and commit hash.
func (l *LockFile) Set(pkg, version, commit string) {
	l.Dependencies[pkg] = LockEntry{
		Version: version,
		Commit:  commit,
	}
}

// Get returns the lock entry for a package, if it exists.
func (l *LockFile) Get(pkg string) (LockEntry, bool) {
	entry, ok := l.Dependencies[pkg]
	return entry, ok
}
