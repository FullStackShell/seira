package shellparse

import (
	"sort"
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func parseStmts(t *testing.T, src string) []*syntax.Stmt {
	t.Helper()
	p := syntax.NewParser(syntax.KeepComments(true))
	f, err := p.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return f.Stmts
}

func TestExtractCallees_DirectCall(t *testing.T) {
	src := `
foo() { echo hello; }
bar() { echo world; }
main() { foo; bar; }
`
	stmts := parseStmts(t, src)
	known := map[string]bool{"foo": true, "bar": true, "main": true}

	// Walk the main function body (last stmt)
	mainStmt := stmts[2]
	callees, hasDynamic := ExtractCallees(mainStmt, known)

	if hasDynamic {
		t.Error("expected no dynamic calls")
	}

	sort.Strings(callees)
	expected := []string{"bar", "foo"}
	if len(callees) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, callees)
	}
	for i, c := range callees {
		if c != expected[i] {
			t.Errorf("callees[%d] = %q, want %q", i, c, expected[i])
		}
	}
}

func TestExtractCallees_CommandWrapper(t *testing.T) {
	src := `main() { command foo; builtin bar; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"foo": true, "bar": true, "main": true}

	callees, hasDynamic := ExtractCallees(stmts[0], known)
	if hasDynamic {
		t.Error("expected no dynamic calls")
	}

	sort.Strings(callees)
	expected := []string{"bar", "foo"}
	if len(callees) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, callees)
	}
}

func TestExtractCallees_CommandSubstitution(t *testing.T) {
	src := `main() { x=$(helper); }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"helper": true, "main": true}

	callees, hasDynamic := ExtractCallees(stmts[0], known)
	if hasDynamic {
		t.Error("expected no dynamic calls")
	}

	if len(callees) != 1 || callees[0] != "helper" {
		t.Errorf("expected [helper], got %v", callees)
	}
}

func TestExtractCallees_DynamicVariable(t *testing.T) {
	src := `main() { $func_name arg1; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"main": true}

	_, hasDynamic := ExtractCallees(stmts[0], known)
	if !hasDynamic {
		t.Error("expected dynamic call detection for $func_name")
	}
}

func TestExtractCallees_Eval(t *testing.T) {
	src := `main() { eval "some_func"; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"main": true, "some_func": true}

	_, hasDynamic := ExtractCallees(stmts[0], known)
	if !hasDynamic {
		t.Error("expected dynamic call detection for eval")
	}
}

func TestExtractCallees_UnknownFunction(t *testing.T) {
	src := `main() { unknown_cmd arg1; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"main": true}

	callees, hasDynamic := ExtractCallees(stmts[0], known)
	if hasDynamic {
		t.Error("expected no dynamic calls")
	}
	if len(callees) != 0 {
		t.Errorf("expected no callees, got %v", callees)
	}
}

func TestExtractCallees_NestedCalls(t *testing.T) {
	src := `main() { if foo; then bar; fi; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"foo": true, "bar": true, "main": true}

	callees, hasDynamic := ExtractCallees(stmts[0], known)
	if hasDynamic {
		t.Error("expected no dynamic calls")
	}

	sort.Strings(callees)
	expected := []string{"bar", "foo"}
	if len(callees) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, callees)
	}
}

func TestExtractCallees_NoDuplicates(t *testing.T) {
	src := `main() { foo; foo; foo; }`
	stmts := parseStmts(t, src)
	known := map[string]bool{"foo": true, "main": true}

	callees, _ := ExtractCallees(stmts[0], known)
	if len(callees) != 1 {
		t.Errorf("expected 1 unique callee, got %v", callees)
	}
}
