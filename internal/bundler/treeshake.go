package bundler

import (
	"log/slog"

	"github.com/Hayao0819/seira/internal/shellparse"
	"mvdan.cc/sh/v3/syntax"
)

// TreeShake filters funcs to only those reachable from the given root
// functions and all side effects. It builds a function-level call graph
// by walking the AST of each function body and effect statement, then
// performs BFS from the roots to find all transitively reachable functions.
//
// If any dynamic (statically unresolvable) call is detected anywhere in the
// reachable code, all functions are returned unchanged as a conservative
// fallback.
//
// roots are the names of functions that serve as entry points (e.g., "main").
// Functions with @seira:keep directives are also treated as roots.
func TreeShake(funcs []funcBlock, effects []effectBlock, roots []string) []funcBlock {
	if len(funcs) == 0 {
		return funcs
	}

	// Build known functions set
	knownFuncs := make(map[string]bool, len(funcs))
	for _, f := range funcs {
		knownFuncs[f.name] = true
	}

	// Build call graph: function name → set of callees
	callGraph := make(map[string][]string)
	globalDynamic := false

	// Map function names to their AST bodies for walking
	funcStmts := make(map[string]*syntax.Stmt, len(funcs))
	for _, f := range funcs {
		funcStmts[f.name] = f.stmt
	}

	// Extract callees from each function body
	for _, f := range funcs {
		callees, hasDynamic := shellparse.ExtractCallees(f.stmt, knownFuncs)
		callGraph[f.name] = callees
		if hasDynamic {
			globalDynamic = true
		}
	}

	// Extract callees from all side effects (these are always reachable)
	effectCallees := make(map[string]bool)
	for _, e := range effects {
		callees, hasDynamic := shellparse.ExtractCallees(e.stmt, knownFuncs)
		for _, c := range callees {
			effectCallees[c] = true
		}
		if hasDynamic {
			globalDynamic = true
		}
	}

	// Collect @seira:keep functions as additional roots
	for _, f := range funcs {
		classified := shellparse.ClassifyStmts([]*syntax.Stmt{f.stmt})
		for _, cs := range classified {
			if cs.Directives.Has("keep") {
				effectCallees[f.name] = true
			}
		}
	}

	// If dynamic calls detected, bail out conservatively
	if globalDynamic {
		slog.Info("tree-shake: dynamic call detected, keeping all functions")
		return funcs
	}

	// Build root set: explicit roots + callees from effects
	reachable := make(map[string]bool)
	queue := make([]string, 0, len(roots)+len(effectCallees))

	for _, r := range roots {
		if knownFuncs[r] {
			queue = append(queue, r)
			reachable[r] = true
		}
	}
	for c := range effectCallees {
		if !reachable[c] {
			queue = append(queue, c)
			reachable[c] = true
		}
	}

	// BFS to find all transitively reachable functions
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, callee := range callGraph[current] {
			if !reachable[callee] {
				reachable[callee] = true
				queue = append(queue, callee)
			}
		}
	}

	// Filter to reachable functions only
	result := make([]funcBlock, 0, len(reachable))
	removed := 0
	for _, f := range funcs {
		if reachable[f.name] {
			result = append(result, f)
		} else {
			removed++
		}
	}

	if removed > 0 {
		slog.Info("tree-shake: removed unused functions", "removed", removed, "kept", len(result))
	}

	return result
}
