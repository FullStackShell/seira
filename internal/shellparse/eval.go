package shellparse

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// EvaluateSourcePath re-evaluates a SourceRef with the given environment variables.
// This is useful when the initial extraction didn't have env context.
func EvaluateSourcePath(ref SourceRef, env map[string]string) SourceRef {
	if !ref.IsDynamic && env == nil {
		return ref
	}
	// Re-parse the raw expression to get AST parts
	parser := syntax.NewParser()
	src := "echo " + ref.Raw
	file, err := parser.Parse(strings.NewReader(src), "")
	if err != nil {
		return ref
	}
	call := file.Stmts[0].Cmd.(*syntax.CallExpr)
	if len(call.Args) < 2 {
		return ref
	}
	return evaluateSourcePathParts(call.Args[1].Parts, env)
}

// evaluateSourcePathParts resolves a source path from AST word parts.
//
// Strategy:
//  1. If all parts are Lit nodes, concatenate directly (fast path).
//  2. For ParamExp with default values (${VAR=default}, ${VAR:-default}, etc.),
//     substitute the default value. If env provides a value for the variable, use that instead.
//  3. If none of the above fully resolve the path, mark IsDynamic=true.
func evaluateSourcePathParts(parts []syntax.WordPart, env map[string]string) SourceRef {
	var raw strings.Builder
	var resolved strings.Builder
	dynamic := false

	printer := syntax.NewPrinter()

	for _, part := range parts {
		// Append to raw representation
		var partRaw strings.Builder
		printer.Print(&partRaw, part)
		raw.WriteString(partRaw.String())

		switch p := part.(type) {
		case *syntax.Lit:
			resolved.WriteString(p.Value)

		case *syntax.DblQuoted:
			// Recursively evaluate the inner parts
			inner := evaluateSourcePathParts(p.Parts, env)
			resolved.WriteString(inner.Resolved)
			if inner.IsDynamic {
				dynamic = true
			}

		case *syntax.ParamExp:
			resolved.WriteString(resolveParamExp(p, env, &dynamic))

		case *syntax.SglQuoted:
			resolved.WriteString(p.Value)

		default:
			// CmdSubst, ArithmExp, etc. — cannot resolve statically
			dynamic = true
		}
	}

	return SourceRef{
		Raw:       raw.String(),
		Resolved:  resolved.String(),
		IsDynamic: dynamic,
	}
}

// resolveParamExp resolves a parameter expansion, using env values or defaults.
func resolveParamExp(p *syntax.ParamExp, env map[string]string, dynamic *bool) string {
	name := ""
	if p.Param != nil {
		name = p.Param.Value
	}

	// Check if env provides a value for this variable
	if env != nil {
		if val, ok := env[name]; ok {
			return val
		}
	}

	// Try to extract default value from expansion operators
	if p.Exp != nil && p.Exp.Word != nil {
		switch p.Exp.Op {
		case syntax.DefaultUnset, // ${VAR-default}
			syntax.DefaultUnsetOrNull, // ${VAR:-default}
			syntax.AssignUnset,        // ${VAR=default}
			syntax.AssignUnsetOrNull:  // ${VAR:=default}
			return wordToString(p.Exp.Word)
		}
	}

	// Cannot resolve statically
	*dynamic = true
	return ""
}

// wordToString extracts the string value from a syntax.Word by concatenating its literal parts.
func wordToString(w *syntax.Word) string {
	var sb strings.Builder
	for _, part := range w.Parts {
		switch p := part.(type) {
		case *syntax.Lit:
			sb.WriteString(p.Value)
		case *syntax.SglQuoted:
			sb.WriteString(p.Value)
		case *syntax.DblQuoted:
			for _, inner := range p.Parts {
				if lit, ok := inner.(*syntax.Lit); ok {
					sb.WriteString(lit.Value)
				}
			}
		}
	}
	return sb.String()
}
