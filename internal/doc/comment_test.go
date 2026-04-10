package doc

import (
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func makeComments(lines ...string) []syntax.Comment {
	var comments []syntax.Comment
	for i, l := range lines {
		comments = append(comments, syntax.Comment{
			Text: " " + l, // simulate "# text" → Comment.Text = " text"
			Hash: syntax.NewPos(0, uint(i+1), 1),
		})
	}
	return comments
}

func TestCollectDocBlock_WithTags(t *testing.T) {
	comments := makeComments(
		"@description This is a function",
		"@param $1 First argument",
	)
	db := collectDocBlock(comments)
	if len(db.lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(db.lines))
	}
}

func TestCollectDocBlock_NoTags(t *testing.T) {
	comments := makeComments(
		"This is just a regular comment",
		"with no tags",
	)
	db := collectDocBlock(comments)
	if len(db.lines) != 0 {
		t.Fatalf("expected 0 lines for non-doc comments, got %d", len(db.lines))
	}
}

func TestParseFuncDoc_Basic(t *testing.T) {
	comments := makeComments(
		"@description Greet someone",
		"@param $1 Name of the person",
		"@param $2 Greeting message",
		"@exitcode 0 Success",
		"@exitcode 1 Missing argument",
		"@return The greeting string",
		"@stdout The greeting message",
		"@see farewell",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "greet", 10)

	if fd.Name != "greet" {
		t.Errorf("expected name 'greet', got %q", fd.Name)
	}
	if fd.Description != "Greet someone" {
		t.Errorf("expected description 'Greet someone', got %q", fd.Description)
	}
	if len(fd.Params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(fd.Params))
	}
	if fd.Params[0].Name != "$1" || fd.Params[0].Description != "Name of the person" {
		t.Errorf("unexpected param[0]: %+v", fd.Params[0])
	}
	if len(fd.ExitCodes) != 2 {
		t.Fatalf("expected 2 exit codes, got %d", len(fd.ExitCodes))
	}
	if fd.Returns != "The greeting string" {
		t.Errorf("unexpected return: %q", fd.Returns)
	}
	if fd.Stdout != "The greeting message" {
		t.Errorf("unexpected stdout: %q", fd.Stdout)
	}
	if len(fd.SeeAlso) != 1 || fd.SeeAlso[0] != "farewell" {
		t.Errorf("unexpected see_also: %v", fd.SeeAlso)
	}
}

func TestParseFuncDoc_MultilineDescription(t *testing.T) {
	comments := makeComments(
		"@description This function does",
		"many things across",
		"multiple lines",
		"@param $1 Arg one",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "multi", 1)

	expected := "This function does\nmany things across\nmultiple lines"
	if fd.Description != expected {
		t.Errorf("expected description %q, got %q", expected, fd.Description)
	}
	if len(fd.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fd.Params))
	}
}

func TestParseFuncDoc_Example(t *testing.T) {
	comments := makeComments(
		"@description A function",
		"@example",
		"  greet World",
		"  echo done",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "greet", 1)

	if len(fd.Examples) != 1 {
		t.Fatalf("expected 1 example, got %d", len(fd.Examples))
	}
	if fd.Examples[0] != "  greet World\n  echo done" {
		t.Errorf("unexpected example: %q", fd.Examples[0])
	}
}

func TestParseFuncDoc_Internal(t *testing.T) {
	comments := makeComments(
		"@description Internal helper",
		"@internal",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "_helper", 1)

	if !fd.Internal {
		t.Error("expected Internal to be true")
	}
}

func TestParseFuncDoc_Options(t *testing.T) {
	comments := makeComments(
		"@description Process files",
		"@option -f --force Force overwrite",
		"@option -v Enable verbose output",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "process", 1)

	if len(fd.Options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(fd.Options))
	}
	if fd.Options[0].Flags != "-f --force" {
		t.Errorf("unexpected flags: %q", fd.Options[0].Flags)
	}
	if fd.Options[0].Description != "Force overwrite" {
		t.Errorf("unexpected description: %q", fd.Options[0].Description)
	}
	if fd.Options[1].Flags != "-v" {
		t.Errorf("unexpected flags: %q", fd.Options[1].Flags)
	}
}

func TestParseFuncDoc_SetVariables(t *testing.T) {
	comments := makeComments(
		"@description Initialize config",
		"@set CONFIG_DIR Path to config directory",
		"@set CONFIG_FILE Path to config file",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "init_config", 1)

	if len(fd.Set) != 2 {
		t.Fatalf("expected 2 set vars, got %d", len(fd.Set))
	}
	if fd.Set[0].Name != "CONFIG_DIR" {
		t.Errorf("unexpected var name: %q", fd.Set[0].Name)
	}
}

func TestParseFileHeader(t *testing.T) {
	comments := makeComments(
		"@file mylib.sh",
		"@brief A utility library",
		"@description This library provides",
		"useful functions for scripting",
		"@set VERSION Library version",
	)
	db := collectDocBlock(comments)
	name, brief, desc, vars := parseFileHeader(db)

	if name != "mylib.sh" {
		t.Errorf("expected name 'mylib.sh', got %q", name)
	}
	if brief != "A utility library" {
		t.Errorf("expected brief 'A utility library', got %q", brief)
	}
	if desc != "This library provides\nuseful functions for scripting" {
		t.Errorf("unexpected description: %q", desc)
	}
	if len(vars) != 1 || vars[0].Name != "VERSION" {
		t.Errorf("unexpected vars: %+v", vars)
	}
}

func TestParseFuncDoc_ArgAlias(t *testing.T) {
	comments := makeComments(
		"@description Test arg alias",
		"@arg $1 First",
		"@arg $2 Second",
	)
	db := collectDocBlock(comments)
	fd := parseFuncDoc(db, "test_func", 1)

	if len(fd.Params) != 2 {
		t.Fatalf("expected 2 params via @arg, got %d", len(fd.Params))
	}
	if fd.Params[0].Name != "$1" {
		t.Errorf("unexpected param name: %q", fd.Params[0].Name)
	}
}
