package lint

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func sampleDiags() []Diagnostic {
	return []Diagnostic{
		{
			Rule:     "bash-source",
			Severity: SeverityWarn,
			File:     "main.sh",
			Line:     5,
			Column:   10,
			Message:  "$BASH_SOURCE will refer to the bundled script, not the original file",
		},
		{
			Rule:     "func-collision",
			Severity: SeverityError,
			File:     "lib/util.sh",
			Line:     12,
			Column:   1,
			Message:  `function "setup" already defined in main.sh:3 — last definition wins`,
		},
	}
}

func TestWriteGCC(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteDiagnostics(&buf, sampleDiags(), FormatGCC); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), buf.String())
	}

	// Check GCC format: file:line:col: severity: message [rule]
	if !strings.HasPrefix(lines[0], "main.sh:5:10: warning:") {
		t.Errorf("unexpected first line: %s", lines[0])
	}
	if !strings.Contains(lines[0], "[bash-source]") {
		t.Errorf("missing rule in first line: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "lib/util.sh:12:1: error:") {
		t.Errorf("unexpected second line: %s", lines[1])
	}
}

func TestWriteJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteDiagnostics(&buf, sampleDiags(), FormatJSON); err != nil {
		t.Fatal(err)
	}

	var parsed []jsonDiagnostic
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	if len(parsed) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(parsed))
	}

	d := parsed[0]
	if d.Rule != "bash-source" || d.Severity != "warn" || d.File != "main.sh" || d.Line != 5 || d.Column != 10 {
		t.Errorf("unexpected first diagnostic: %+v", d)
	}
}

func TestWriteJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteDiagnostics(&buf, []Diagnostic{}, FormatJSON); err != nil {
		t.Fatal(err)
	}

	var parsed []jsonDiagnostic
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(parsed) != 0 {
		t.Errorf("expected empty array, got %d items", len(parsed))
	}
}

func TestWriteText(t *testing.T) {
	var buf bytes.Buffer
	if err := WriteDiagnostics(&buf, sampleDiags(), FormatText); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	if !strings.Contains(out, "⚠") {
		t.Error("expected warning icon in text output")
	}
	if !strings.Contains(out, "✗") {
		t.Error("expected error icon in text output")
	}
	if !strings.Contains(out, "2 issue(s) found.") {
		t.Error("expected issue count in text output")
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input string
		want  Format
		err   bool
	}{
		{"text", FormatText, false},
		{"gcc", FormatGCC, false},
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"", FormatText, false},
		{"xml", "", true},
	}

	for _, tt := range tests {
		got, err := ParseFormat(tt.input)
		if (err != nil) != tt.err {
			t.Errorf("ParseFormat(%q): err=%v, wantErr=%v", tt.input, err, tt.err)
		}
		if got != tt.want {
			t.Errorf("ParseFormat(%q)=%q, want %q", tt.input, got, tt.want)
		}
	}
}
