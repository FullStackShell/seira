package shellparse

import (
	"mvdan.cc/sh/v3/syntax"
)

// StmtKind represents the classification of a top-level statement.
type StmtKind int

const (
	StmtFunc   StmtKind = iota // function definition
	StmtSource                 // source or . command
	StmtEffect                 // side effect (variable assignment, command execution, etc.)
)

// ClassifiedStmt is a top-level statement with its classification.
type ClassifiedStmt struct {
	Kind       StmtKind
	Stmt       *syntax.Stmt
	FuncName   string // set when Kind == StmtFunc
	SourcePath string // set when Kind == StmtSource (raw expression)
}

// ClassifyStmts classifies top-level statements of a script.
func ClassifyStmts(stmts []*syntax.Stmt) []ClassifiedStmt {
	var result []ClassifiedStmt
	for _, stmt := range stmts {
		result = append(result, classifyStmt(stmt))
	}
	return result
}

func classifyStmt(stmt *syntax.Stmt) ClassifiedStmt {
	// Check for function declaration
	if fn, ok := stmt.Cmd.(*syntax.FuncDecl); ok {
		return ClassifiedStmt{
			Kind:     StmtFunc,
			Stmt:     stmt,
			FuncName: fn.Name.Value,
		}
	}

	// Check for source/. command
	if call, ok := stmt.Cmd.(*syntax.CallExpr); ok && len(call.Args) >= 2 {
		if lit, ok := call.Args[0].Parts[0].(*syntax.Lit); ok {
			if lit.Value == "source" || lit.Value == "." {
				// Extract the source path (raw, for reference)
				raw := ""
				if len(call.Args) >= 2 {
					printer := syntax.NewPrinter()
					var buf []byte
					w := &byteWriter{buf: &buf}
					printer.Print(w, call.Args[1])
					raw = string(buf)
				}
				return ClassifiedStmt{
					Kind:       StmtSource,
					Stmt:       stmt,
					SourcePath: raw,
				}
			}
		}
	}

	// Everything else is a side effect
	return ClassifiedStmt{
		Kind: StmtEffect,
		Stmt: stmt,
	}
}

// byteWriter is a simple io.Writer that appends to a byte slice.
type byteWriter struct {
	buf *[]byte
}

func (w *byteWriter) Write(p []byte) (int, error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
