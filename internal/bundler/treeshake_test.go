package bundler

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func parseFuncBlock(t *testing.T, name, body string) funcBlock {
	t.Helper()
	src := name + "() { " + body + " }"
	p := syntax.NewParser(syntax.KeepComments(true))
	f, err := p.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatalf("parse error for %s: %v", name, err)
	}
	return funcBlock{
		origin: "test.sh",
		name:   name,
		stmt:   f.Stmts[0],
	}
}

func parseEffectBlock(t *testing.T, src string) effectBlock {
	t.Helper()
	p := syntax.NewParser(syntax.KeepComments(true))
	f, err := p.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	return effectBlock{
		origin: "test.sh",
		stmt:   f.Stmts[0],
	}
}

func funcNames(blocks []funcBlock) []string {
	names := make([]string, len(blocks))
	for i, b := range blocks {
		names[i] = b.name
	}
	return names
}

func TestTreeShake_RemovesUnused(t *testing.T) {
	funcs := []funcBlock{
		parseFuncBlock(t, "main", "helper1;"),
		parseFuncBlock(t, "helper1", "echo ok;"),
		parseFuncBlock(t, "unused", "echo unused;"),
	}

	result := TreeShake(funcs, nil, []string{"main"})
	names := funcNames(result)

	if len(names) != 2 {
		t.Fatalf("expected 2 functions, got %v", names)
	}
	for _, n := range names {
		if n == "unused" {
			t.Error("unused function should have been removed")
		}
	}
}

func TestTreeShake_TransitiveDeps(t *testing.T) {
	funcs := []funcBlock{
		parseFuncBlock(t, "main", "a;"),
		parseFuncBlock(t, "a", "b;"),
		parseFuncBlock(t, "b", "echo end;"),
		parseFuncBlock(t, "orphan", "echo orphan;"),
	}

	result := TreeShake(funcs, nil, []string{"main"})
	names := funcNames(result)

	expected := map[string]bool{"main": true, "a": true, "b": true}
	if len(names) != len(expected) {
		t.Fatalf("expected %d functions, got %v", len(expected), names)
	}
	for _, n := range names {
		if !expected[n] {
			t.Errorf("unexpected function %q in result", n)
		}
	}
}

func TestTreeShake_EffectsAsRoots(t *testing.T) {
	funcs := []funcBlock{
		parseFuncBlock(t, "setup", "echo setup;"),
		parseFuncBlock(t, "unused", "echo unused;"),
	}
	effects := []effectBlock{
		parseEffectBlock(t, "setup"),
	}

	result := TreeShake(funcs, effects, nil)
	names := funcNames(result)

	if len(names) != 1 || names[0] != "setup" {
		t.Errorf("expected [setup], got %v", names)
	}
}

func TestTreeShake_DynamicFallback(t *testing.T) {
	funcs := []funcBlock{
		parseFuncBlock(t, "main", "$dynamic_call;"),
		parseFuncBlock(t, "unused", "echo unused;"),
	}

	result := TreeShake(funcs, nil, []string{"main"})

	// Dynamic call detected → all functions should be kept
	if len(result) != len(funcs) {
		t.Errorf("expected all %d functions (dynamic fallback), got %d", len(funcs), len(result))
	}
}

func TestTreeShake_EmptyFuncs(t *testing.T) {
	result := TreeShake(nil, nil, []string{"main"})
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestTreeShake_PreservesOrder(t *testing.T) {
	funcs := []funcBlock{
		parseFuncBlock(t, "z_first", "echo z;"),
		parseFuncBlock(t, "main", "z_first; a_last;"),
		parseFuncBlock(t, "a_last", "echo a;"),
		parseFuncBlock(t, "unused", "echo unused;"),
	}

	result := TreeShake(funcs, nil, []string{"main"})
	names := funcNames(result)

	// Order should be preserved from original
	expected := []string{"z_first", "main", "a_last"}
	if len(names) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, names)
	}
	for i, n := range names {
		if n != expected[i] {
			t.Errorf("result[%d] = %q, want %q", i, n, expected[i])
		}
	}
}
