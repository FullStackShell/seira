package shellparse

import (
	"mvdan.cc/sh/v3/syntax"
)

// ExtractCallees walks the AST rooted at node and returns the names of
// known functions that appear in call position, plus whether any dynamic
// (statically unresolvable) call was detected.
//
// Dynamic calls include: variable-based calls ($func), eval with arguments,
// and other patterns where the callee cannot be determined statically.
// When hasDynamic is true, callers should conservatively assume that any
// function might be called.
func ExtractCallees(node syntax.Node, knownFuncs map[string]bool) (callees []string, hasDynamic bool) {
	seen := map[string]bool{}

	syntax.Walk(node, func(n syntax.Node) bool {
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}

		firstWord := call.Args[0]

		// Check for dynamic call: first word contains parameter expansion
		if containsParamExp(firstWord) {
			hasDynamic = true
			return true
		}

		name := literalValue(firstWord)
		if name == "" {
			return true
		}

		// eval with arguments → dynamic
		if name == "eval" && len(call.Args) > 1 {
			hasDynamic = true
			return true
		}

		// "command funcname" or "builtin funcname" → unwrap one level
		if (name == "command" || name == "builtin") && len(call.Args) >= 2 {
			inner := literalValue(call.Args[1])
			if inner != "" && knownFuncs[inner] && !seen[inner] {
				seen[inner] = true
				callees = append(callees, inner)
			}
			return true
		}

		// Direct function call
		if knownFuncs[name] && !seen[name] {
			seen[name] = true
			callees = append(callees, name)
		}

		return true
	})

	return callees, hasDynamic
}

// literalValue returns the string value of a word if it consists entirely
// of literal parts (possibly concatenated). Returns "" if the word contains
// any non-literal parts (expansions, substitutions, etc.).
func literalValue(word *syntax.Word) string {
	var s string
	for _, part := range word.Parts {
		lit, ok := part.(*syntax.Lit)
		if !ok {
			return ""
		}
		s += lit.Value
	}
	return s
}

// containsParamExp returns true if the word contains any parameter expansion,
// indicating a dynamically-determined value.
func containsParamExp(word *syntax.Word) bool {
	for _, part := range word.Parts {
		switch part.(type) {
		case *syntax.ParamExp:
			return true
		case *syntax.CmdSubst:
			return true
		case *syntax.ArithmExp:
			return true
		}
		// DblQuoted may contain expansions
		if dq, ok := part.(*syntax.DblQuoted); ok {
			for _, inner := range dq.Parts {
				switch inner.(type) {
				case *syntax.ParamExp:
					return true
				case *syntax.CmdSubst:
					return true
				}
			}
		}
	}
	return false
}
