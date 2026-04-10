package doc

import (
	"bytes"
	"strings"
	"testing"
)

func sampleProjectDoc() ProjectDoc {
	return ProjectDoc{
		Title: "Test Project",
		Files: []FileDoc{
			{
				Path:        "lib/util.sh",
				Name:        "util.sh",
				Brief:       "Utility functions",
				Description: "A collection of useful utilities.",
				Variables: []Variable{
					{Name: "VERSION", Description: "Library version"},
				},
				Functions: []FuncDoc{
					{
						Name:        "greet",
						Description: "Say hello to someone",
						Params: []Param{
							{Name: "$1", Description: "Person's name"},
						},
						ExitCodes: []ExitCode{
							{Code: "0", Description: "Success"},
						},
						Examples: []string{"greet World"},
						SeeAlso:  []string{"farewell"},
						Line:     10,
					},
					{
						Name:        "_internal_helper",
						Description: "Internal function",
						Internal:    true,
						Line:        20,
					},
				},
				Sections: []Section{
					{
						Name: "String Functions",
						Functions: []FuncDoc{
							{
								Name:        "str_upper",
								Description: "Convert to uppercase",
								Params: []Param{
									{Name: "$1", Description: "Input string"},
								},
								Stdout: "Uppercased string",
								Line:   30,
							},
						},
					},
				},
			},
		},
	}
}

func TestRenderMarkdown(t *testing.T) {
	pd := sampleProjectDoc()
	var buf bytes.Buffer
	if err := RenderMarkdown(&buf, pd); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	// Single file in project: no multi-file headers, uses # for title, ## for file, ### for func
	checks := []string{
		"# Test Project",
		"# util.sh",
		"Utility functions",
		"A collection of useful utilities.",
		"## greet()",
		"Say hello to someone",
		"`$1`",
		"Person's name",
		"`0`",
		"Success",
		"```bash",
		"greet World",
		"**See Also:** farewell",
		"## String Functions",
		"### str_upper()",
		"**VERSION**",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("markdown output missing %q", check)
		}
	}

	// Internal function should be hidden
	if strings.Contains(out, "_internal_helper") {
		t.Error("internal function should not appear in markdown output")
	}
}

func TestRenderHTML(t *testing.T) {
	pd := sampleProjectDoc()
	var buf bytes.Buffer
	if err := RenderHTML(&buf, pd); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	checks := []string{
		"<!DOCTYPE html>",
		"<title>Test Project</title>",
		"<h1>Test Project</h1>",
		"util.sh",
		"greet()",
		"Say hello to someone",
		"<code>$1</code>",
		"String Functions",
		"str_upper()",
		"<pre><code>greet World</code></pre>",
		"<code>VERSION</code>",
	}

	for _, check := range checks {
		if !strings.Contains(out, check) {
			t.Errorf("HTML output missing %q", check)
		}
	}

	// Internal function should be hidden
	if strings.Contains(out, "_internal_helper") {
		t.Error("internal function should not appear in HTML output")
	}
}

func TestRenderMarkdownFile_SingleFile(t *testing.T) {
	fd := FileDoc{
		Path: "test.sh",
		Name: "test.sh",
		Functions: []FuncDoc{
			{
				Name:        "hello",
				Description: "Say hello",
				Line:        1,
			},
		},
	}
	var buf bytes.Buffer
	if err := RenderMarkdownFile(&buf, fd); err != nil {
		t.Fatal(err)
	}
	out := buf.String()

	// Single file uses h1 for file and h2 for functions
	if !strings.Contains(out, "# test.sh") {
		t.Error("single file should use h1")
	}
	if !strings.Contains(out, "## hello()") {
		t.Error("single file function should use h2")
	}
}
