package bpkg

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Hayao0819/seira/internal/config"
	"github.com/cockroachdb/errors"
)

const defaultRemote = "https://raw.githubusercontent.com"

// PackageRef represents a parsed package reference like "user/name@version".
type PackageRef struct {
	User    string
	Name    string
	Version string // git tag or branch; empty means "master"
}

// ParsePackageRef parses a string like "user/name@version" into a PackageRef.
func ParsePackageRef(s string) (*PackageRef, error) {
	ref := &PackageRef{Version: "master"}

	// Split version
	if idx := strings.LastIndex(s, "@"); idx > 0 {
		ref.Version = s[idx+1:]
		s = s[:idx]
	}

	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("invalid package reference %q: must be user/name[@version]", s)
	}
	ref.User = parts[0]
	ref.Name = parts[1]

	return ref, nil
}

// String returns the canonical string form of the package reference.
func (r *PackageRef) String() string {
	if r.Version != "" && r.Version != "master" {
		return fmt.Sprintf("%s/%s@%s", r.User, r.Name, r.Version)
	}
	return fmt.Sprintf("%s/%s", r.User, r.Name)
}

// Installer manages package installation.
type Installer struct {
	DepsDir string // target deps directory (default: ./deps)
	Remote  string // raw content remote URL
}

// NewInstaller creates a new Installer with sensible defaults.
func NewInstaller(depsDir string) *Installer {
	if depsDir == "" {
		depsDir = "deps"
	}
	return &Installer{
		DepsDir: depsDir,
		Remote:  defaultRemote,
	}
}

// InstallResult holds metadata about an installed package.
type InstallResult struct {
	Ref    *PackageRef
	Commit string            // resolved git commit hash
	Sub    []*InstallResult  // results from transitive dependencies
}

// PkgKey returns the canonical "user/name" key for this package.
func (r *InstallResult) PkgKey() string {
	return fmt.Sprintf("%s/%s", r.Ref.User, r.Ref.Name)
}

// Install downloads and installs a package locally.
// It first tries bpkg-compatible install (manifest-based), then falls back to
// git clone for repositories without bpkg.json.
// After installation, if the package is a seira project with dependencies,
// those are installed recursively.
func (inst *Installer) Install(ref *PackageRef) (*InstallResult, error) {
	slog.Info("installing package", "package", ref.String())

	var commit string
	var err error

	// Try bpkg-compatible install first
	manifest, mErr := inst.fetchManifest(ref)
	if mErr == nil {
		commit, err = inst.installWithManifest(ref, manifest)
	} else {
		// Fallback: git clone for non-bpkg repositories
		slog.Info("no bpkg manifest found, falling back to git clone", "package", ref.String())
		commit, err = inst.installWithGitClone(ref)
	}
	if err != nil {
		return nil, err
	}

	result := &InstallResult{Ref: ref, Commit: commit}

	// Resolve seira project dependencies recursively
	pkgDir := filepath.Join(inst.DepsDir, ref.Name)
	subResults, err := inst.resolveSeiraDepsFull(pkgDir)
	if err != nil {
		return nil, errors.Wrapf(err, "resolving seira dependencies for %s", ref.String())
	}
	result.Sub = subResults

	return result, nil
}

// installWithManifest installs a package using its bpkg manifest.
// Returns the resolved git commit hash.
func (inst *Installer) installWithManifest(ref *PackageRef, manifest *Manifest) (string, error) {
	pkgName := manifest.Name
	if pkgName == "" {
		pkgName = ref.Name
	}

	pkgDir := filepath.Join(inst.DepsDir, pkgName)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return "", errors.Wrap(err, "creating package directory")
	}

	// Download scripts
	for _, script := range manifest.Scripts {
		if err := inst.downloadFile(ref, script, pkgDir); err != nil {
			return "", errors.Wrapf(err, "downloading script %s", script)
		}
		if err := os.Chmod(filepath.Join(pkgDir, script), 0755); err != nil {
			return "", errors.Wrapf(err, "setting permissions for %s", script)
		}
	}

	// Download additional files
	for _, file := range manifest.Files {
		if err := inst.downloadFile(ref, file, pkgDir); err != nil {
			return "", errors.Wrapf(err, "downloading file %s", file)
		}
	}

	// Write manifest to package directory
	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", errors.Wrap(err, "marshaling manifest")
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "bpkg.json"), append(manifestData, '\n'), 0644); err != nil {
		return "", errors.Wrap(err, "writing manifest")
	}

	// Create symlinks in deps/bin/
	if err := inst.createBinSymlinks(pkgName, manifest.Scripts); err != nil {
		return "", err
	}

	// Install transitive bpkg dependencies
	if len(manifest.Dependencies) > 0 {
		slog.Info("installing transitive dependencies", "package", ref.String(), "count", len(manifest.Dependencies))
		for pkg, ver := range manifest.Dependencies {
			depRef, err := ParsePackageRef(pkg)
			if err != nil {
				return "", errors.Wrapf(err, "parsing dependency %s", pkg)
			}
			if ver != "" && ver != "*" {
				depRef.Version = ver
			}
			if _, err := inst.Install(depRef); err != nil {
				return "", errors.Wrapf(err, "installing dependency %s", pkg)
			}
		}
	}

	// Resolve commit hash via GitHub API
	commit, err := inst.fetchCommitHash(ref)
	if err != nil {
		slog.Warn("could not resolve commit hash", "package", ref.String(), "error", err)
		commit = ""
	}

	slog.Info("installed package (bpkg)", "package", ref.String(), "dir", pkgDir)
	return commit, nil
}

// installWithGitClone installs a repository that doesn't have a bpkg manifest.
// It clones the repository into a temp directory, then copies shell scripts and
// relevant files to the deps directory. Returns the resolved commit hash.
func (inst *Installer) installWithGitClone(ref *PackageRef) (string, error) {
	cloneURL := fmt.Sprintf("https://github.com/%s/%s.git", ref.User, ref.Name)

	// Clone to temp directory
	tmpDir, err := os.MkdirTemp("", "seira-install-*")
	if err != nil {
		return "", errors.Wrap(err, "creating temp directory")
	}
	defer os.RemoveAll(tmpDir)

	cloneArgs := []string{"clone", "--depth", "1"}
	if ref.Version != "master" {
		cloneArgs = append(cloneArgs, "--branch", ref.Version)
	}
	cloneArgs = append(cloneArgs, cloneURL, tmpDir)

	cmd := exec.Command("git", cloneArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		// If branch-based clone fails for "master", also try "main"
		if ref.Version == "master" {
			// Remove the previously failed tmpDir and create fresh
			os.RemoveAll(tmpDir)
			os.MkdirAll(tmpDir, 0755)
			cmd2 := exec.Command("git", "clone", "--depth", "1", "--branch", "main", cloneURL, tmpDir)
			if out2, err2 := cmd2.CombinedOutput(); err2 != nil {
				return "", errors.Wrapf(err, "git clone failed: %s\nAlso tried 'main': %s", out, out2)
			}
		} else {
			return "", errors.Wrapf(err, "git clone failed: %s", out)
		}
	}

	// Get commit hash before removing .git
	commit := getGitHeadCommit(tmpDir)

	pkgDir := filepath.Join(inst.DepsDir, ref.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return "", errors.Wrap(err, "creating package directory")
	}

	// Copy all files except .git directory
	if err := copyDirContents(tmpDir, pkgDir); err != nil {
		return "", errors.Wrap(err, "copying repository contents")
	}

	// Find shell scripts and create bin symlinks
	scripts, err := findShellScripts(pkgDir)
	if err != nil {
		return "", errors.Wrap(err, "finding shell scripts")
	}
	if err := inst.createBinSymlinks(ref.Name, scripts); err != nil {
		return "", err
	}

	slog.Info("installed package (git clone)", "package", ref.String(), "dir", pkgDir)
	return commit, nil
}

// getGitHeadCommit returns the HEAD commit hash for a git repository.
func getGitHeadCommit(repoDir string) string {
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// fetchCommitHash resolves the commit hash for a package version via GitHub API.
func (inst *Installer) fetchCommitHash(ref *PackageRef) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s",
		ref.User, ref.Name, ref.Version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d for %s", resp.StatusCode, url)
	}

	var result struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.SHA, nil
}

// resolveSeiraDepsFull checks if an installed package is a seira project
// and recursively installs its dependencies.
func (inst *Installer) resolveSeiraDepsFull(pkgDir string) ([]*InstallResult, error) {
	cfg, err := config.Load(pkgDir)
	if err != nil {
		return nil, nil // not a seira project or config error, skip
	}

	if len(cfg.Dependencies) == 0 {
		return nil, nil
	}

	slog.Info("resolving seira project dependencies", "dir", pkgDir, "count", len(cfg.Dependencies))

	var results []*InstallResult
	for pkg, ver := range cfg.Dependencies {
		depRef, err := ParsePackageRef(pkg)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing seira dependency %s", pkg)
		}
		if ver != "" && ver != "*" {
			depRef.Version = ver
		}
		result, err := inst.Install(depRef)
		if err != nil {
			return nil, errors.Wrapf(err, "installing seira dependency %s", pkg)
		}
		results = append(results, result)
	}
	return results, nil
}

// createBinSymlinks creates symlinks in deps/bin/ for the given scripts.
func (inst *Installer) createBinSymlinks(pkgName string, scripts []string) error {
	if len(scripts) == 0 {
		return nil
	}

	binDir := filepath.Join(inst.DepsDir, "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return errors.Wrap(err, "creating bin directory")
	}

	for _, script := range scripts {
		linkName := strings.TrimSuffix(filepath.Base(script), ".sh")
		linkPath := filepath.Join(binDir, linkName)
		target := filepath.Join("..", pkgName, script)

		os.Remove(linkPath)
		if err := os.Symlink(target, linkPath); err != nil {
			return errors.Wrapf(err, "creating symlink for %s", script)
		}
	}
	return nil
}

// copyDirContents copies all files and subdirectories from src to dst,
// excluding .git directory. Symlinks are recreated as symlinks.
func copyDirContents(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.Name() == ".git" {
			continue
		}

		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// Handle symlinks
		if entry.Type()&os.ModeSymlink != 0 {
			target, err := os.Readlink(srcPath)
			if err != nil {
				return err
			}
			os.Remove(dstPath)
			if err := os.Symlink(target, dstPath); err != nil {
				return err
			}
			continue
		}

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0755); err != nil {
				return err
			}
			if err := copyDirContents(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if err := os.WriteFile(dstPath, data, info.Mode()); err != nil {
				return err
			}
		}
	}
	return nil
}

// findShellScripts finds all .sh files in dir (relative paths).
func findShellScripts(dir string) ([]string, error) {
	var scripts []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(info.Name(), ".sh") {
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			scripts = append(scripts, rel)
		}
		return nil
	})
	return scripts, err
}

// fetchManifest downloads and parses the bpkg.json for a package.
func (inst *Installer) fetchManifest(ref *PackageRef) (*Manifest, error) {
	// Try bpkg.json first, then package.json
	for _, name := range []string{"bpkg.json", "package.json"} {
		url := fmt.Sprintf("%s/%s/%s/%s/%s", inst.Remote, ref.User, ref.Name, ref.Version, name)
		slog.Debug("fetching manifest", "url", url)

		resp, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrap(err, "HTTP request failed")
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "reading response body")
		}

		var m Manifest
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, errors.Wrapf(err, "parsing %s", name)
		}
		return &m, nil
	}

	return nil, fmt.Errorf("no bpkg.json or package.json found for %s", ref)
}

// downloadFile downloads a single file from the package repository.
func (inst *Installer) downloadFile(ref *PackageRef, filePath string, destDir string) error {
	url := fmt.Sprintf("%s/%s/%s/%s/%s", inst.Remote, ref.User, ref.Name, ref.Version, filePath)
	slog.Debug("downloading file", "url", url)

	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrap(err, "HTTP request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}

	destPath := filepath.Join(destDir, filePath)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return errors.Wrap(err, "creating directory")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading response body")
	}

	return os.WriteFile(destPath, data, 0644)
}
