package shellparse

import (
	"strings"
	"testing"
)

func TestClassifyStmts(t *testing.T) {
	src := `#!/bin/bash
MY_VAR="hello"
source lib/helper.sh
greet() { echo "Hello"; }
echo "side effect"
. ./other.sh
main() { greet; }
`
	parser := NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	classified := ClassifyStmts(file.Stmts)

	expected := []struct {
		kind     StmtKind
		funcName string
	}{
		{StmtEffect, ""},       // MY_VAR="hello"
		{StmtSource, ""},       // source lib/helper.sh
		{StmtFunc, "greet"},    // greet() { ... }
		{StmtEffect, ""},       // echo "side effect"
		{StmtSource, ""},       // . ./other.sh
		{StmtFunc, "main"},     // main() { ... }
	}

	if len(classified) != len(expected) {
		t.Fatalf("expected %d statements, got %d", len(expected), len(classified))
	}

	for i, exp := range expected {
		got := classified[i]
		if got.Kind != exp.kind {
			t.Errorf("stmt[%d]: kind got %d, want %d", i, got.Kind, exp.kind)
		}
		if exp.funcName != "" && got.FuncName != exp.funcName {
			t.Errorf("stmt[%d]: funcName got %q, want %q", i, got.FuncName, exp.funcName)
		}
	}
}

func TestClassifyStmts_SourcePath(t *testing.T) {
	src := `source lib/helper.sh`
	parser := NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	classified := ClassifyStmts(file.Stmts)
	if len(classified) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(classified))
	}
	if classified[0].Kind != StmtSource {
		t.Errorf("expected StmtSource, got %d", classified[0].Kind)
	}
	if !strings.Contains(classified[0].SourcePath, "lib/helper.sh") {
		t.Errorf("SourcePath should contain 'lib/helper.sh', got %q", classified[0].SourcePath)
	}
}
