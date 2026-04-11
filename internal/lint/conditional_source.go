package lint

import (
	"mvdan.cc/sh/v3/syntax"
)

// ConditionalSourceRule detects source/. commands inside if/case/while/for blocks.
// In concat mode, all sources are resolved at build time — conditional sourcing
// is silently flattened, which may change runtime behavior.
type ConditionalSourceRule struct{}

func (r *ConditionalSourceRule) Name() string { return "conditional-source" }

func (r *ConditionalSourceRule) Check(ctx *Context) []Diagnostic {
	if ctx.Mode != "concat" {
		return nil
	}

	var diags []Diagnostic
	for _, path := range ctx.Order {
		node := ctx.Graph.Node(path)
		if node == nil || node.Script == nil {
			continue
		}
		rel := relPath(ctx.BaseDir, path)

		for _, stmt := range node.Script.File.Stmts {
			diags = append(diags, r.walkStmt(stmt, rel, 0)...)
		}
	}
	return diags
}

// walkStmt recursively checks for source commands nested inside control flow.
func (r *ConditionalSourceRule) walkStmt(stmt *syntax.Stmt, file string, depth int) []Diagnostic {
	if stmt.Cmd == nil {
		return nil
	}

	var diags []Diagnostic

	switch cmd := stmt.Cmd.(type) {
	case *syntax.IfClause:
		diags = append(diags, r.walkIfClause(cmd, file, depth)...)

	case *syntax.WhileClause:
		for _, s := range cmd.Do {
			diags = append(diags, r.walkStmt(s, file, depth+1)...)
		}

	case *syntax.ForClause:
		for _, s := range cmd.Do {
			diags = append(diags, r.walkStmt(s, file, depth+1)...)
		}

	case *syntax.CaseClause:
		for _, item := range cmd.Items {
			for _, s := range item.Stmts {
				diags = append(diags, r.walkStmt(s, file, depth+1)...)
			}
		}

	case *syntax.FuncDecl:
		// source inside a function body is called at runtime — skip
		return nil

	case *syntax.CallExpr:
		if depth > 0 && isSourceCall(cmd) {
			diags = append(diags, Diagnostic{
				Rule:     r.Name(),
				Severity: SeverityWarn,
				File:     file,
				Line:     int(stmt.Pos().Line()),
				Column:   int(stmt.Pos().Col()),
				Message:  "source inside conditional block is resolved at build time in concat mode",
			})
		}

	case *syntax.BinaryCmd:
		diags = append(diags, r.walkStmt(cmd.X, file, depth)...)
		diags = append(diags, r.walkStmt(cmd.Y, file, depth)...)
	}

	return diags
}

// walkIfClause handles if/elif/else chains.
func (r *ConditionalSourceRule) walkIfClause(ic *syntax.IfClause, file string, depth int) []Diagnostic {
	var diags []Diagnostic
	for _, s := range ic.Then {
		diags = append(diags, r.walkStmt(s, file, depth+1)...)
	}
	if ic.Else != nil {
		diags = append(diags, r.walkIfClause(ic.Else, file, depth)...)
	}
	return diags
}

// isSourceCall checks if a CallExpr is a source/. command.
func isSourceCall(call *syntax.CallExpr) bool {
	if len(call.Args) < 2 || len(call.Args[0].Parts) == 0 {
		return false
	}
	lit, ok := call.Args[0].Parts[0].(*syntax.Lit)
	if !ok {
		return false
	}
	return lit.Value == "source" || lit.Value == "."
}
