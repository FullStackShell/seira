package shellparse

import (
	"strings"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func TestParseDirectives_Basic(t *testing.T) {
	comments := []syntax.Comment{
		{Text: " @seira:ignore"},
	}
	directives := ParseDirectives(comments)
	if len(directives) != 1 {
		t.Fatalf("expected 1 directive, got %d", len(directives))
	}
	if directives[0].Name != "ignore" {
		t.Errorf("expected name 'ignore', got %q", directives[0].Name)
	}
	if directives[0].Args != "" {
		t.Errorf("expected empty args, got %q", directives[0].Args)
	}
}

func TestParseDirectives_WithArgs(t *testing.T) {
	comments := []syntax.Comment{
		{Text: " @seira:type string"},
		{Text: " @seira:param name string"},
	}
	directives := ParseDirectives(comments)
	if len(directives) != 2 {
		t.Fatalf("expected 2 directives, got %d", len(directives))
	}
	if directives[0].Name != "type" || directives[0].Args != "string" {
		t.Errorf("directive 0: got %q %q", directives[0].Name, directives[0].Args)
	}
	if directives[1].Name != "param" || directives[1].Args != "name string" {
		t.Errorf("directive 1: got %q %q", directives[1].Name, directives[1].Args)
	}
}

func TestParseDirectives_IgnoresNonSeira(t *testing.T) {
	comments := []syntax.Comment{
		{Text: " regular comment"},
		{Text: " @seira:ignore"},
		{Text: " another comment"},
	}
	directives := ParseDirectives(comments)
	if len(directives) != 1 {
		t.Fatalf("expected 1 directive, got %d", len(directives))
	}
}

func TestParseDirectives_Empty(t *testing.T) {
	directives := ParseDirectives(nil)
	if len(directives) != 0 {
		t.Fatalf("expected 0 directives, got %d", len(directives))
	}
}

func TestDirectives_Has(t *testing.T) {
	d := Directives{
		{Name: "ignore"},
		{Name: "type", Args: "string"},
	}
	if !d.Has("ignore") {
		t.Error("expected Has('ignore') to be true")
	}
	if !d.Has("type") {
		t.Error("expected Has('type') to be true")
	}
	if d.Has("unknown") {
		t.Error("expected Has('unknown') to be false")
	}
}

func TestDirectives_Get(t *testing.T) {
	d := Directives{
		{Name: "type", Args: "string"},
	}
	dir, ok := d.Get("type")
	if !ok {
		t.Fatal("expected to find 'type' directive")
	}
	if dir.Args != "string" {
		t.Errorf("expected args 'string', got %q", dir.Args)
	}
	_, ok = d.Get("missing")
	if ok {
		t.Error("expected not to find 'missing' directive")
	}
}

func TestDirectives_GetAll(t *testing.T) {
	d := Directives{
		{Name: "param", Args: "name string"},
		{Name: "param", Args: "age number"},
		{Name: "return", Args: "string"},
	}
	params := d.GetAll("param")
	if len(params) != 2 {
		t.Fatalf("expected 2 param directives, got %d", len(params))
	}
}

func TestParseDirectives_FromParsedScript(t *testing.T) {
	src := `#!/bin/bash
# @seira:ignore
source /etc/config.sh

source lib/utils.sh
`
	parser := NewParser()
	script, err := parser.Analyze(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	if len(script.Sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(script.Sources))
	}

	// First source should have ignore directive
	if !script.Sources[0].Directives.Has("ignore") {
		t.Error("first source should have 'ignore' directive")
	}

	// Second source should have no directives
	if len(script.Sources[1].Directives) != 0 {
		t.Error("second source should have no directives")
	}
}
