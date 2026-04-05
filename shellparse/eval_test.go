package shellparse

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func parseWord(t *testing.T, s string) *syntax.Word {
	t.Helper()
	// Parse as a command: echo <word> to extract the word
	src := "echo " + s
	parser := syntax.NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test")
	if err != nil {
		t.Fatalf("failed to parse %q: %v", s, err)
	}
	call := file.Stmts[0].Cmd.(*syntax.CallExpr)
	return call.Args[1] // the word after "echo"
}

func TestEvaluateSourcePath_Literal(t *testing.T) {
	w := parseWord(t, "lib/hoge.sh")
	ref := evaluateSourcePathParts(w.Parts, nil)
	if ref.Resolved != "lib/hoge.sh" {
		t.Errorf("Resolved: got %q, want %q", ref.Resolved, "lib/hoge.sh")
	}
	if ref.IsDynamic {
		t.Error("expected IsDynamic=false for literal path")
	}
}

func TestEvaluateSourcePath_DefaultValue(t *testing.T) {
	// ${SEIRA_ROOTDIR="."}/lib/hoge.sh
	w := parseWord(t, `"${SEIRA_ROOTDIR="."}/lib/hoge.sh"`)
	ref := evaluateSourcePathParts(w.Parts, nil)
	if ref.Resolved != "./lib/hoge.sh" {
		t.Errorf("Resolved: got %q, want %q", ref.Resolved, "./lib/hoge.sh")
	}
	if ref.IsDynamic {
		t.Error("expected IsDynamic=false when default value is available")
	}
}

func TestEvaluateSourcePath_ColonDefault(t *testing.T) {
	// ${DIR:-/opt}/lib.sh
	w := parseWord(t, `"${DIR:-/opt}/lib.sh"`)
	ref := evaluateSourcePathParts(w.Parts, nil)
	if ref.Resolved != "/opt/lib.sh" {
		t.Errorf("Resolved: got %q, want %q", ref.Resolved, "/opt/lib.sh")
	}
	if ref.IsDynamic {
		t.Error("expected IsDynamic=false")
	}
}

func TestEvaluateSourcePath_EnvOverride(t *testing.T) {
	w := parseWord(t, `"${SEIRA_ROOTDIR="."}/lib/hoge.sh"`)
	env := map[string]string{"SEIRA_ROOTDIR": "/custom"}
	ref := evaluateSourcePathParts(w.Parts, env)
	if ref.Resolved != "/custom/lib/hoge.sh" {
		t.Errorf("Resolved: got %q, want %q", ref.Resolved, "/custom/lib/hoge.sh")
	}
}

func TestEvaluateSourcePath_Dynamic(t *testing.T) {
	// $1 — purely dynamic, cannot resolve
	w := parseWord(t, `"$1"`)
	ref := evaluateSourcePathParts(w.Parts, nil)
	if !ref.IsDynamic {
		t.Error("expected IsDynamic=true for positional parameter")
	}
}

func TestEvaluateSourcePath_SimpleVar(t *testing.T) {
	// $FOO — no default, no env → dynamic
	w := parseWord(t, `"$FOO"`)
	ref := evaluateSourcePathParts(w.Parts, nil)
	if !ref.IsDynamic {
		t.Error("expected IsDynamic=true for unresolvable variable")
	}
}

func TestEvaluateSourcePath_SimpleVarWithEnv(t *testing.T) {
	// $FOO with env → resolves
	w := parseWord(t, `"$FOO"`)
	env := map[string]string{"FOO": "/path/to/lib.sh"}
	ref := evaluateSourcePathParts(w.Parts, env)
	if ref.Resolved != "/path/to/lib.sh" {
		t.Errorf("Resolved: got %q, want %q", ref.Resolved, "/path/to/lib.sh")
	}
	if ref.IsDynamic {
		t.Error("expected IsDynamic=false when env provides the value")
	}
}
