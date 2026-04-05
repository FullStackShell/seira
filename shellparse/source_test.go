package shellparse

import (
	"strings"
	"testing"
)

func TestExtractSources_Literal(t *testing.T) {
	src := `#!/bin/bash
source lib/hoge.sh
. lib/fuga.sh
`
	parser := NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}
	refs := extractSources(file)
	if len(refs) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(refs))
	}
	if refs[0].Resolved != "lib/hoge.sh" {
		t.Errorf("refs[0].Resolved: got %q, want %q", refs[0].Resolved, "lib/hoge.sh")
	}
	if refs[1].Resolved != "lib/fuga.sh" {
		t.Errorf("refs[1].Resolved: got %q, want %q", refs[1].Resolved, "lib/fuga.sh")
	}
}

func TestExtractSources_WithVariable(t *testing.T) {
	src := `#!/bin/bash
source "${DIR="."}/lib.sh"
`
	parser := NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}
	refs := extractSources(file)
	if len(refs) != 1 {
		t.Fatalf("expected 1 source, got %d", len(refs))
	}
	if refs[0].Resolved != "./lib.sh" {
		t.Errorf("Resolved: got %q, want %q", refs[0].Resolved, "./lib.sh")
	}
}

func TestExtractSources_DotCommand(t *testing.T) {
	src := `. ./helpers.sh`
	parser := NewParser()
	file, err := parser.Parse(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}
	refs := extractSources(file)
	if len(refs) != 1 {
		t.Fatalf("expected 1 source, got %d", len(refs))
	}
	if refs[0].Resolved != "./helpers.sh" {
		t.Errorf("Resolved: got %q, want %q", refs[0].Resolved, "./helpers.sh")
	}
}
