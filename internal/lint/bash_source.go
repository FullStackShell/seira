package lint

import (
	"mvdan.cc/sh/v3/syntax"
)

// BashSourceRule detects usage of BASH_SOURCE and $0 for path resolution.
// These break in concat mode because all code runs as a single script.
type BashSourceRule struct{}

func (r *BashSourceRule) Name() string { return "bash-source" }

func (r *BashSourceRule) Check(ctx *Context) []Diagnostic {
	if ctx.Mode != "concat" && ctx.Mode != "library" {
		return nil
	}

	var diags []Diagnostic
	for _, path := range ctx.Order {
		node := ctx.Graph.Node(path)
		if node == nil || node.Script == nil {
			continue
		}
		rel := relPath(ctx.BaseDir, path)

		syntax.Walk(node.Script.File, func(n syntax.Node) bool {
			pe, ok := n.(*syntax.ParamExp)
			if !ok || pe.Param == nil {
				return true
			}

			name := pe.Param.Value
			line := int(pe.Pos().Line())

			switch {
			case name == "BASH_SOURCE":
				diags = append(diags, Diagnostic{
					Rule:     r.Name(),
					Severity: SeverityWarn,
					File:     rel,
					Line:     line,
					Message:  "$BASH_SOURCE will refer to the bundled script, not the original file",
				})
			case name == "0" && hasSuffixRemoval(pe):
				// ${0%/*} pattern — used for dirname-like path resolution
				diags = append(diags, Diagnostic{
					Rule:     r.Name(),
					Severity: SeverityWarn,
					File:     rel,
					Line:     line,
					Message:  "$0 with path manipulation will refer to the bundled script, not the original file",
				})
			}
			return true
		})
	}
	return diags
}

// hasSuffixRemoval checks whether a ParamExp uses % or %% operators.
func hasSuffixRemoval(pe *syntax.ParamExp) bool {
	if pe.Exp == nil {
		return false
	}
	op := pe.Exp.Op
	return op == syntax.RemSmallSuffix || op == syntax.RemLargeSuffix
}
