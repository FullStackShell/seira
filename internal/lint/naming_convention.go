package lint

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/Hayao0819/seira/internal/shellparse"
	"mvdan.cc/sh/v3/syntax"
)

// NamingConventionRule validates that function names and global variable
// declarations follow the project's declared prefix convention.
// This rule is opt-in: it only activates when ctx.Prefix is non-empty.
//
// Functions must start with one of the declared prefixes followed by the
// separator ("::" for bash, "_" for sh). OOP class methods (starting with
// an uppercase letter followed by "::") are exempt.
//
// Global variables (declare -g/-gA/-gi or top-level assignments) must
// match _SEIRA_<PREFIX>_* (internal) or SEIRA_<PREFIX>_* (public).
type NamingConventionRule struct{}

func (r *NamingConventionRule) Name() string { return "naming-convention" }

func (r *NamingConventionRule) Check(ctx *Context) []Diagnostic {
	if len(ctx.Prefix) == 0 {
		return nil
	}

	sep := "::"
	if ctx.Shell == "sh" {
		sep = "_"
	}

	// Build allowed function prefixes
	funcPrefixes := make([]string, 0, len(ctx.Prefix)*2)
	for _, p := range ctx.Prefix {
		funcPrefixes = append(funcPrefixes, p+sep)
		if sep == "_" {
			// POSIX sh: internal functions use _prefix_
			funcPrefixes = append(funcPrefixes, "_"+p+"_")
		} else {
			// bash: internal functions use prefix::_
			funcPrefixes = append(funcPrefixes, p+"::_")
		}
	}

	// Build allowed global variable prefixes
	// Allow both exact match (_SEIRA_OOP) and prefixed (_SEIRA_OOP_NEXT_ID)
	varPrefixes := make([]string, 0, len(ctx.Prefix)*4)
	for _, p := range ctx.Prefix {
		upper := strings.ToUpper(p)
		varPrefixes = append(varPrefixes, "_SEIRA_"+upper+"_") // _SEIRA_OOP_*
		varPrefixes = append(varPrefixes, "SEIRA_"+upper+"_")  // SEIRA_OOP_*
	}
	varExact := make([]string, 0, len(ctx.Prefix)*2)
	for _, p := range ctx.Prefix {
		upper := strings.ToUpper(p)
		varExact = append(varExact, "_SEIRA_"+upper) // _SEIRA_OOP
		varExact = append(varExact, "SEIRA_"+upper)  // SEIRA_OOP
	}

	var diags []Diagnostic

	for _, path := range ctx.Order {
		node := ctx.Graph.Node(path)
		if node == nil || node.Script == nil {
			continue
		}
		rel := relPath(ctx.BaseDir, path)

		// Check function names
		classified := shellparse.ClassifyStmts(node.Script.File.Stmts)
		for _, cs := range classified {
			if cs.Kind == shellparse.StmtFunc {
				name := cs.FuncName
				line := int(cs.Stmt.Pos().Line())
				col := int(cs.Stmt.Pos().Col())

				if !isValidFuncName(name, funcPrefixes, sep) {
					diags = append(diags, Diagnostic{
						Rule:     r.Name(),
						Severity: SeverityWarn,
						File:     rel,
						Line:     line,
						Column:   col,
						Message:  fmt.Sprintf("function %q does not match declared prefix %v", name, ctx.Prefix),
					})
				}
			}
		}

		// Check global variable declarations
		diags = append(diags, r.checkGlobalVars(node.Script.File, rel, varPrefixes, varExact)...)
	}

	return diags
}

// isValidFuncName checks if a function name matches any of the allowed prefixes
// or is an OOP class method (UpperCaseName::method).
func isValidFuncName(name string, prefixes []string, sep string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}

	// OOP class methods: ClassName::method where ClassName starts with uppercase.
	// Only applicable for bash (:: separator).
	if sep == "::" && strings.Contains(name, "::") {
		className := name[:strings.Index(name, "::")]
		if len(className) > 0 && unicode.IsUpper(rune(className[0])) {
			return true
		}
	}

	return false
}

// checkGlobalVars walks the AST to find global variable declarations
// (declare -g/-gA/-gi) and top-level assignments, then validates their names.
func (r *NamingConventionRule) checkGlobalVars(file *syntax.File, relFile string, varPrefixes []string, varExact []string) []Diagnostic {
	var diags []Diagnostic

	for _, stmt := range file.Stmts {
		// Check declare -g style declarations
		if decl, ok := stmt.Cmd.(*syntax.DeclClause); ok {
			if !isGlobalDeclare(decl) {
				continue
			}
			for _, assign := range decl.Args {
				if assign.Name == nil {
					continue
				}
				varName := assign.Name.Value
				if !isValidVarName(varName, varPrefixes, varExact) {
					diags = append(diags, Diagnostic{
						Rule:     r.Name(),
						Severity: SeverityWarn,
						File:     relFile,
						Line:     int(assign.Name.Pos().Line()),
						Column:   int(assign.Name.Pos().Col()),
						Message:  fmt.Sprintf("global variable %q does not match prefix pattern (expected _SEIRA_<PREFIX>_* or SEIRA_<PREFIX>_*)", varName),
					})
				}
			}
		}

		// Check top-level simple assignments (VAR=value at file scope)
		if call, ok := stmt.Cmd.(*syntax.CallExpr); ok && len(call.Args) == 0 && len(call.Assigns) > 0 {
			for _, assign := range call.Assigns {
				if assign.Name == nil {
					continue
				}
				varName := assign.Name.Value
				if !isValidVarName(varName, varPrefixes, varExact) {
					diags = append(diags, Diagnostic{
						Rule:     r.Name(),
						Severity: SeverityWarn,
						File:     relFile,
						Line:     int(assign.Name.Pos().Line()),
						Column:   int(assign.Name.Pos().Col()),
						Message:  fmt.Sprintf("global variable %q does not match prefix pattern (expected _SEIRA_<PREFIX>_* or SEIRA_<PREFIX>_*)", varName),
					})
				}
			}
		}
	}

	return diags
}

// isGlobalDeclare checks if a DeclClause is a global declaration (declare -g, -gA, -gi, etc.).
func isGlobalDeclare(decl *syntax.DeclClause) bool {
	if decl.Variant == nil {
		return false
	}
	if decl.Variant.Value != "declare" {
		return false
	}
	for _, arg := range decl.Args {
		if arg.Name == nil && arg.Value != nil {
			// This is a flag like -gA
			val := arg.Value.Parts
			for _, part := range val {
				if lit, ok := part.(*syntax.Lit); ok {
					if strings.Contains(lit.Value, "g") {
						return true
					}
				}
			}
		}
	}
	return false
}

// isValidVarName checks if a variable name matches any of the allowed prefixes or exact names.
func isValidVarName(name string, prefixes []string, exact ...[]string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(name, p) {
			return true
		}
	}
	for _, ex := range exact {
		for _, e := range ex {
			if name == e {
				return true
			}
		}
	}
	return false
}
