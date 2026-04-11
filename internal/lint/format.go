package lint

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Format represents an output format for diagnostics.
type Format string

const (
	FormatText Format = "text" // human-readable (default)
	FormatGCC  Format = "gcc"  // GCC-compatible: file:line:col: severity: message [rule]
	FormatJSON Format = "json" // JSON array of diagnostics
)

// ParseFormat parses a format string, returning an error for unknown formats.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "text", "":
		return FormatText, nil
	case "gcc":
		return FormatGCC, nil
	case "json":
		return FormatJSON, nil
	default:
		return "", fmt.Errorf("unknown format %q (supported: text, gcc, json)", s)
	}
}

// WriteDiagnostics writes diagnostics in the specified format to the writer.
func WriteDiagnostics(w io.Writer, diags []Diagnostic, format Format) error {
	switch format {
	case FormatJSON:
		return writeJSON(w, diags)
	case FormatGCC:
		return writeGCC(w, diags)
	default:
		return writeText(w, diags)
	}
}

func writeText(w io.Writer, diags []Diagnostic) error {
	if len(diags) == 0 {
		_, err := fmt.Fprintln(w, "No issues found.")
		return err
	}
	for _, d := range diags {
		icon := "⚠"
		if d.Severity == SeverityError {
			icon = "✗"
		}
		loc := d.File
		if d.Line > 0 {
			loc = fmt.Sprintf("%s:%d", d.File, d.Line)
		}
		if _, err := fmt.Fprintf(w, "  %s %s: %s [%s]\n", icon, loc, d.Message, d.Rule); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintf(w, "\n%d issue(s) found.\n", len(diags))
	return err
}

func writeGCC(w io.Writer, diags []Diagnostic) error {
	for _, d := range diags {
		sev := "warning"
		if d.Severity == SeverityError {
			sev = "error"
		}

		line := d.Line
		if line == 0 {
			line = 1
		}
		col := d.Column
		if col == 0 {
			col = 1
		}

		if _, err := fmt.Fprintf(w, "%s:%d:%d: %s: %s [%s]\n", d.File, line, col, sev, d.Message, d.Rule); err != nil {
			return err
		}
	}
	return nil
}

type jsonDiagnostic struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Message  string `json:"message"`
}

func writeJSON(w io.Writer, diags []Diagnostic) error {
	out := make([]jsonDiagnostic, len(diags))
	for i, d := range diags {
		out[i] = jsonDiagnostic{
			Rule:     d.Rule,
			Severity: d.Severity.String(),
			File:     d.File,
			Line:     d.Line,
			Column:   d.Column,
			Message:  d.Message,
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
