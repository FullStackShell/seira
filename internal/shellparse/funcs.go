package shellparse

import "mvdan.cc/sh/v3/syntax"

// extractTopLevelFuncs extracts function declarations at the top level of a script.
func extractTopLevelFuncs(stmts []*syntax.Stmt) []syntax.FuncDecl {
	var funcs []syntax.FuncDecl
	for _, stmt := range stmts {
		if fn, ok := stmt.Cmd.(*syntax.FuncDecl); ok {
			funcs = append(funcs, *fn)
		}
	}
	return funcs
}
