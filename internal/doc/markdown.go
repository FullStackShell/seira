package doc

import (
	"fmt"
	"io"
	"strings"
)

// RenderMarkdown writes the project documentation in Markdown format.
func RenderMarkdown(w io.Writer, pd ProjectDoc) error {
	mw := &mdWriter{w: w}

	if pd.Title != "" {
		mw.writef("# %s\n\n", pd.Title)
	}

	// Table of contents
	if len(pd.Files) > 1 {
		mw.writef("## Table of Contents\n\n")
		for _, f := range pd.Files {
			title := f.Name
			if title == "" {
				title = f.Path
			}
			mw.writef("- [%s](#%s)\n", title, tocAnchor(title))
		}
		mw.writef("\n---\n\n")
	}

	for i, f := range pd.Files {
		if err := renderMarkdownFile(mw, f, len(pd.Files) > 1); err != nil {
			return err
		}
		if i < len(pd.Files)-1 {
			mw.writef("\n---\n\n")
		}
	}

	return mw.err
}

// RenderMarkdownFile writes documentation for a single file.
func RenderMarkdownFile(w io.Writer, fd FileDoc) error {
	mw := &mdWriter{w: w}
	if err := renderMarkdownFile(mw, fd, false); err != nil {
		return err
	}
	return mw.err
}

func renderMarkdownFile(mw *mdWriter, fd FileDoc, multiFile bool) error {
	prefix := "##"
	funcPrefix := "###"
	if !multiFile {
		prefix = "#"
		funcPrefix = "##"
	}

	// File header
	title := fd.Name
	if title == "" {
		title = fd.Path
	}
	mw.writef("%s %s\n\n", prefix, title)

	if fd.Brief != "" {
		mw.writef("%s\n\n", fd.Brief)
	}
	if fd.Description != "" {
		mw.writef("%s\n\n", fd.Description)
	}

	// File-level variables
	if len(fd.Variables) > 0 {
		mw.writef("%s# Global Variables\n\n", string(funcPrefix[0]))
		for _, v := range fd.Variables {
			mw.writef("- **%s**: %s\n", v.Name, v.Description)
		}
		mw.writef("\n")
	}

	// Functions (not in a section)
	for _, fn := range fd.Functions {
		renderMarkdownFunc(mw, fn, funcPrefix)
	}

	// Sections
	for _, sec := range fd.Sections {
		mw.writef("%s %s\n\n", funcPrefix, sec.Name)
		sectionFuncPrefix := funcPrefix + "#"
		for _, fn := range sec.Functions {
			renderMarkdownFunc(mw, fn, sectionFuncPrefix)
		}
	}

	return nil
}

func renderMarkdownFunc(mw *mdWriter, fn FuncDoc, prefix string) {
	if fn.Internal {
		return // skip internal functions
	}

	mw.writef("%s %s()\n\n", prefix, fn.Name)

	if fn.Description != "" {
		mw.writef("%s\n\n", fn.Description)
	}

	// Parameters
	if len(fn.Params) > 0 {
		mw.writef("**Arguments:**\n\n")
		for _, p := range fn.Params {
			mw.writef("- `%s` — %s\n", p.Name, p.Description)
		}
		mw.writef("\n")
	}

	// Options
	if len(fn.Options) > 0 {
		mw.writef("**Options:**\n\n")
		for _, o := range fn.Options {
			mw.writef("- `%s` — %s\n", o.Flags, o.Description)
		}
		mw.writef("\n")
	}

	// Stdin
	if fn.Stdin != "" {
		mw.writef("**Stdin:** %s\n\n", fn.Stdin)
	}

	// Stdout
	if fn.Stdout != "" {
		mw.writef("**Stdout:** %s\n\n", fn.Stdout)
	}

	// Return
	if fn.Returns != "" {
		mw.writef("**Returns:** %s\n\n", fn.Returns)
	}

	// Exit codes
	if len(fn.ExitCodes) > 0 {
		mw.writef("**Exit Codes:**\n\n")
		for _, ec := range fn.ExitCodes {
			mw.writef("- `%s` — %s\n", ec.Code, ec.Description)
		}
		mw.writef("\n")
	}

	// Set variables
	if len(fn.Set) > 0 {
		mw.writef("**Sets:**\n\n")
		for _, v := range fn.Set {
			mw.writef("- `%s` — %s\n", v.Name, v.Description)
		}
		mw.writef("\n")
	}

	// Examples
	for _, ex := range fn.Examples {
		mw.writef("**Example:**\n\n")
		mw.writef("```bash\n%s\n```\n\n", strings.TrimSpace(ex))
	}

	// See also
	if len(fn.SeeAlso) > 0 {
		mw.writef("**See Also:** %s\n\n", strings.Join(fn.SeeAlso, ", "))
	}
}

// tocAnchor generates a GitHub-compatible markdown anchor from a title.
func tocAnchor(title string) string {
	s := strings.ToLower(title)
	s = strings.ReplaceAll(s, " ", "-")
	// Remove characters that aren't alphanumeric, hyphen, or underscore
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

type mdWriter struct {
	w   io.Writer
	err error
}

func (mw *mdWriter) writef(format string, args ...any) {
	if mw.err != nil {
		return
	}
	_, mw.err = fmt.Fprintf(mw.w, format, args...)
}
