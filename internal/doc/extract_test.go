package doc

import (
	"strings"
	"testing"

	"github.com/Hayao0819/seira/internal/shellparse"
)

func TestExtractFile_BasicFunction(t *testing.T) {
	src := `#!/bin/bash
# @file test.sh
# @brief Test script

# @description Say hello to someone
# @param $1 Name
# @exitcode 0 Success
greet() {
    echo "Hello, $1"
}
`
	parser := shellparse.NewParser()
	script, err := parser.Analyze(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	fd := ExtractFile(script, ".")
	if fd.Name != "test.sh" {
		t.Errorf("expected file name 'test.sh', got %q", fd.Name)
	}
	if fd.Brief != "Test script" {
		t.Errorf("expected brief 'Test script', got %q", fd.Brief)
	}
	if len(fd.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(fd.Functions))
	}
	fn := fd.Functions[0]
	if fn.Name != "greet" {
		t.Errorf("expected function name 'greet', got %q", fn.Name)
	}
	if fn.Description != "Say hello to someone" {
		t.Errorf("expected description 'Say hello to someone', got %q", fn.Description)
	}
	if len(fn.Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(fn.Params))
	}
}

func TestExtractFile_Sections(t *testing.T) {
	src := `#!/bin/bash

# @section String Functions

# @description Convert to uppercase
str_upper() {
    echo "${1^^}"
}

# @description Convert to lowercase
str_lower() {
    echo "${1,,}"
}

# @section Math Functions

# @description Add two numbers
math_add() {
    echo $(( $1 + $2 ))
}
`
	parser := shellparse.NewParser()
	script, err := parser.Analyze(strings.NewReader(src), "lib.sh")
	if err != nil {
		t.Fatal(err)
	}

	fd := ExtractFile(script, ".")
	if len(fd.Sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(fd.Sections))
	}
	if fd.Sections[0].Name != "String Functions" {
		t.Errorf("expected section name 'String Functions', got %q", fd.Sections[0].Name)
	}
	if len(fd.Sections[0].Functions) != 2 {
		t.Errorf("expected 2 functions in first section, got %d", len(fd.Sections[0].Functions))
	}
	if fd.Sections[1].Name != "Math Functions" {
		t.Errorf("expected section name 'Math Functions', got %q", fd.Sections[1].Name)
	}
	if len(fd.Sections[1].Functions) != 1 {
		t.Errorf("expected 1 function in second section, got %d", len(fd.Sections[1].Functions))
	}
}

func TestExtractFile_InternalHidden(t *testing.T) {
	src := `#!/bin/bash

# @description Public function
public_func() {
    _helper
}

# @description Internal helper
# @internal
_helper() {
    echo "internal"
}
`
	parser := shellparse.NewParser()
	script, err := parser.Analyze(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	fd := ExtractFile(script, ".")
	if len(fd.Functions) != 2 {
		t.Fatalf("expected 2 functions, got %d", len(fd.Functions))
	}

	// Verify internal flag
	for _, fn := range fd.Functions {
		if fn.Name == "_helper" && !fn.Internal {
			t.Error("expected _helper to be internal")
		}
		if fn.Name == "public_func" && fn.Internal {
			t.Error("expected public_func to not be internal")
		}
	}
}

func TestExtractFile_NoDocFunctions(t *testing.T) {
	src := `#!/bin/bash
# No doc tags here
no_doc() {
    echo "no doc"
}
`
	parser := shellparse.NewParser()
	script, err := parser.Analyze(strings.NewReader(src), "test.sh")
	if err != nil {
		t.Fatal(err)
	}

	fd := ExtractFile(script, ".")
	// Function should be present but with empty documentation
	if len(fd.Functions) != 1 {
		t.Fatalf("expected 1 function, got %d", len(fd.Functions))
	}
	if fd.Functions[0].Description != "" {
		t.Errorf("expected empty description for undocumented func, got %q", fd.Functions[0].Description)
	}
}
