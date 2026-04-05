package shellparse

import (
	"mvdan.cc/sh/v3/syntax"
)

// extractSources walks the AST and returns SourceRef for each source/. command.
func extractSources(file *syntax.File) []SourceRef {
	var refs []SourceRef

	syntax.Walk(file, func(n syntax.Node) bool {
		stmt, ok := n.(*syntax.Stmt)
		if !ok {
			return true
		}

		directives := ParseDirectives(stmt.Comments)

		for _, call := range findCmdCalls(stmt, "source", ".") {
			// Arguments after the command name are the sourced files
			for _, arg := range call.Args[1:] {
				ref := evaluateSourcePathParts(arg.Parts, nil)
				ref.Directives = directives
				refs = append(refs, ref)
			}
		}
		return true
	})

	return refs
}

// findCmdCalls finds all CallExpr nodes within a node that invoke one of the given commands.
func findCmdCalls(node syntax.Node, cmds ...string) []syntax.CallExpr {
	cmdSet := make(map[string]struct{}, len(cmds))
	for _, c := range cmds {
		cmdSet[c] = struct{}{}
	}

	var results []syntax.CallExpr
	syntax.Walk(node, func(n syntax.Node) bool {
		call, ok := n.(*syntax.CallExpr)
		if !ok || len(call.Args) == 0 || len(call.Args[0].Parts) == 0 {
			return true
		}
		lit, ok := call.Args[0].Parts[0].(*syntax.Lit)
		if !ok {
			return true
		}
		if _, match := cmdSet[lit.Value]; match {
			results = append(results, *call)
		}
		return true
	})

	return results
}
