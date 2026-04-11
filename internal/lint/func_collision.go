package lint

import (
	"fmt"

	"github.com/Hayao0819/seira/internal/shellparse"
)

// FuncCollisionRule detects duplicate function names across different files.
// In concat mode, all functions share a single namespace — last definition wins,
// which may silently override earlier ones.
type FuncCollisionRule struct{}

func (r *FuncCollisionRule) Name() string { return "func-collision" }

func (r *FuncCollisionRule) Check(ctx *Context) []Diagnostic {
	if ctx.Mode != "concat" && ctx.Mode != "library" {
		return nil
	}

	type funcLoc struct {
		file string
		line int
	}
	seen := map[string]funcLoc{}

	var diags []Diagnostic
	for _, path := range ctx.Order {
		node := ctx.Graph.Node(path)
		if node == nil || node.Script == nil {
			continue
		}
		rel := relPath(ctx.BaseDir, path)

		classified := shellparse.ClassifyStmts(node.Script.File.Stmts)
		for _, cs := range classified {
			if cs.Kind != shellparse.StmtFunc {
				continue
			}
			name := cs.FuncName
			line := int(cs.Stmt.Pos().Line())
			col := int(cs.Stmt.Pos().Col())

			if prev, exists := seen[name]; exists && prev.file != rel {
				diags = append(diags, Diagnostic{
					Rule:     r.Name(),
					Severity: SeverityWarn,
					File:     rel,
					Line:     line,
					Column:   col,
					Message:  fmt.Sprintf("function %q already defined in %s:%d — last definition wins", name, prev.file, prev.line),
				})
			} else if !exists {
				seen[name] = funcLoc{file: rel, line: line}
			}
		}
	}
	return diags
}
