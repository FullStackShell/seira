package lint

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/depgraph"
	"github.com/Hayao0819/seira/internal/shellparse"
	"github.com/cockroachdb/errors"
)

// Severity represents the importance level of a diagnostic.
type Severity int

const (
	SeverityWarn  Severity = iota // potential issue, may still work
	SeverityError                 // will break at runtime
)

func (s Severity) String() string {
	switch s {
	case SeverityWarn:
		return "warn"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// Diagnostic is a single lint finding.
type Diagnostic struct {
	Rule     string   // rule identifier, e.g. "bash-source"
	Severity Severity // warn or error
	File     string   // relative file path
	Line     int      // 1-based line number (0 if unknown)
	Column   int      // 1-based column number (0 if unknown)
	Message  string   // human-readable description
}

func (d Diagnostic) String() string {
	loc := d.File
	if d.Line > 0 && d.Column > 0 {
		loc = fmt.Sprintf("%s:%d:%d", d.File, d.Line, d.Column)
	} else if d.Line > 0 {
		loc = fmt.Sprintf("%s:%d", d.File, d.Line)
	}
	return fmt.Sprintf("[%s] %s: %s (%s)", d.Severity, loc, d.Message, d.Rule)
}

// Context provides the information rules need to analyze.
type Context struct {
	Graph   *depgraph.Graph
	Order   []string // topologically sorted paths (absolute)
	BaseDir string   // absolute base directory for relative paths
	Mode    string   // "concat", "tarball", "library"
	Prefix  []string // function namespace prefixes (from config, opt-in for naming rule)
	Shell   string   // "bash" or "sh" (determines naming convention separator)
}

// Rule is a single lint check.
type Rule interface {
	Name() string
	Check(ctx *Context) []Diagnostic
}

// Runner executes a set of rules and collects diagnostics.
type Runner struct {
	rules []Rule
}

// NewRunner creates a runner with the default rule set.
func NewRunner() *Runner {
	return &Runner{
		rules: []Rule{
			&BashSourceRule{},
			&ConditionalSourceRule{},
			&FuncCollisionRule{},
			&NamingConventionRule{},
		},
	}
}

// Run executes all rules and returns diagnostics.
func (r *Runner) Run(ctx *Context) []Diagnostic {
	var all []Diagnostic
	for _, rule := range r.rules {
		all = append(all, rule.Check(ctx)...)
	}
	return all
}

// LogDiagnostics emits diagnostics via slog.
func LogDiagnostics(diags []Diagnostic) {
	for _, d := range diags {
		loc := d.File
		if d.Line > 0 {
			loc = fmt.Sprintf("%s:%d", d.File, d.Line)
		}
		switch d.Severity {
		case SeverityError:
			slog.Error(d.Message, "rule", d.Rule, "location", loc)
		default:
			slog.Warn(d.Message, "rule", d.Rule, "location", loc)
		}
	}
}

// Config holds settings for running the linter.
type Config struct {
	InputPath string
	BaseDir   string
	Mode      string            // bundle mode to check against
	Type      string            // "executable" or "library"
	Env       map[string]string // environment variables for source resolution
	Prefix    []string          // function namespace prefixes for naming convention lint
	Shell     string            // "bash" or "sh" (determines naming convention separator)
}

// Lint runs all lint rules on the given input and returns diagnostics.
func Lint(cfg Config) ([]Diagnostic, error) {
	absInput, err := filepath.Abs(cfg.InputPath)
	if err != nil {
		return nil, errors.Wrap(err, "resolving input path")
	}

	baseDir := cfg.BaseDir
	if baseDir == "" {
		baseDir = filepath.Dir(absInput)
	}
	baseDir, err = filepath.Abs(baseDir)
	if err != nil {
		return nil, errors.Wrap(err, "resolving base dir")
	}

	parser := shellparse.NewParser()
	graph, err := depgraph.Resolve(parser, absInput, cfg.Env)
	if err != nil {
		return nil, errors.Wrap(err, "resolving dependencies")
	}

	order, err := graph.TopologicalSort()
	if err != nil {
		return nil, errors.Wrap(err, "topological sort")
	}

	mode := cfg.Mode
	if cfg.Type == "library" {
		mode = "library"
	}

	ctx := &Context{
		Graph:   graph,
		Order:   order,
		BaseDir: baseDir,
		Mode:    mode,
		Prefix:  cfg.Prefix,
		Shell:   cfg.Shell,
	}

	return NewRunner().Run(ctx), nil
}

// relPath returns a relative path from base to target, falling back to target.
func relPath(base, target string) string {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return target
	}
	return rel
}
