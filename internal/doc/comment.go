package doc

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// docBlock holds raw comment lines belonging to one documentation block.
type docBlock struct {
	lines []string
	pos   syntax.Pos
}

// collectDocBlock extracts the doc-comment block from the comments attached
// to a statement. Only contiguous comment lines immediately preceding the
// statement that contain at least one @-tag (or are part of such a block)
// are included.
func collectDocBlock(comments []syntax.Comment) docBlock {
	if len(comments) == 0 {
		return docBlock{}
	}

	// Collect all comment texts, stripping the leading "# " or "#".
	lines := make([]string, 0, len(comments))
	for _, c := range comments {
		lines = append(lines, strings.TrimPrefix(c.Text, " "))
	}

	// Check if any line has a doc tag — if not, this isn't a doc block.
	hasDocTag := false
	for _, l := range lines {
		if isDocTag(l) {
			hasDocTag = true
			break
		}
	}
	if !hasDocTag {
		return docBlock{}
	}

	pos := comments[0].Hash
	return docBlock{lines: lines, pos: pos}
}

// isToolDirective returns true if the line is a tool directive comment
// (e.g. shellcheck, noinspection) that should not be treated as doc content.
func isToolDirective(line string) bool {
	trimmed := strings.TrimSpace(line)
	prefixes := []string{
		"shellcheck ",
		"noinspection ",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

// isDocTag returns true if the line starts with a recognized documentation tag.
func isDocTag(line string) bool {
	trimmed := strings.TrimSpace(line)
	tags := []string{
		"@description", "@param", "@arg",
		"@return", "@exitcode", "@example",
		"@see", "@internal", "@set",
		"@option", "@stdin", "@stdout",
		"@file", "@brief", "@section",
		"@noargs",
	}
	for _, tag := range tags {
		if strings.HasPrefix(trimmed, tag) {
			return true
		}
	}
	return false
}

// parseFuncDoc parses a docBlock into a FuncDoc.
func parseFuncDoc(db docBlock, funcName string, line int) FuncDoc {
	fd := FuncDoc{
		Name: funcName,
		Line: line,
	}
	if len(db.lines) == 0 {
		return fd
	}

	var descLines []string
	inDescription := false
	inExample := false
	var exampleLines []string

	for _, rawLine := range db.lines {
		trimmed := strings.TrimSpace(rawLine)

		// Handle @example blocks
		if inExample {
			if isDocTag(trimmed) && !strings.HasPrefix(trimmed, "@example") {
				// End of example block, save accumulated example
				fd.Examples = append(fd.Examples, strings.Join(exampleLines, "\n"))
				exampleLines = nil
				inExample = false
				// Fall through to process this tag
			} else if isToolDirective(trimmed) {
				continue // skip tool directives
			} else if trimmed == "" || !isDocTag(trimmed) {
				exampleLines = append(exampleLines, rawLine)
				continue
			} else {
				// Another @example tag — save current and start new
				fd.Examples = append(fd.Examples, strings.Join(exampleLines, "\n"))
				exampleLines = nil
				// Fall through to process @example
			}
		}

		// Handle multi-line @description continuation
		if inDescription {
			if isDocTag(trimmed) {
				inDescription = false
				// Fall through to process this tag
			} else if isToolDirective(trimmed) {
				continue // skip tool directives
			} else {
				descLines = append(descLines, rawLine)
				continue
			}
		}

		switch {
		case strings.HasPrefix(trimmed, "@description"):
			inDescription = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@description"))
			if rest != "" {
				descLines = append(descLines, rest)
			}

		case strings.HasPrefix(trimmed, "@param") || strings.HasPrefix(trimmed, "@arg"):
			tag := "@param"
			if strings.HasPrefix(trimmed, "@arg") {
				tag = "@arg"
			}
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, tag))
			name, desc, _ := strings.Cut(rest, " ")
			fd.Params = append(fd.Params, Param{
				Name:        name,
				Description: strings.TrimSpace(desc),
			})

		case strings.HasPrefix(trimmed, "@option"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@option"))
			// Options can be "-f" or "-f --force" followed by description
			parts := strings.Fields(rest)
			var flags []string
			var descParts []string
			inFlags := true
			for _, p := range parts {
				if inFlags && strings.HasPrefix(p, "-") {
					flags = append(flags, p)
				} else {
					inFlags = false
					descParts = append(descParts, p)
				}
			}
			fd.Options = append(fd.Options, Option{
				Flags:       strings.Join(flags, " "),
				Description: strings.Join(descParts, " "),
			})

		case strings.HasPrefix(trimmed, "@exitcode"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@exitcode"))
			code, desc, _ := strings.Cut(rest, " ")
			fd.ExitCodes = append(fd.ExitCodes, ExitCode{
				Code:        code,
				Description: strings.TrimSpace(desc),
			})

		case strings.HasPrefix(trimmed, "@return"):
			fd.Returns = strings.TrimSpace(strings.TrimPrefix(trimmed, "@return"))

		case strings.HasPrefix(trimmed, "@stdin"):
			fd.Stdin = strings.TrimSpace(strings.TrimPrefix(trimmed, "@stdin"))

		case strings.HasPrefix(trimmed, "@stdout"):
			fd.Stdout = strings.TrimSpace(strings.TrimPrefix(trimmed, "@stdout"))

		case strings.HasPrefix(trimmed, "@example"):
			inExample = true
			// If there's content on the same line as @example, treat it as the first line
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@example"))
			if rest != "" {
				exampleLines = append(exampleLines, rest)
			}

		case strings.HasPrefix(trimmed, "@see"):
			ref := strings.TrimSpace(strings.TrimPrefix(trimmed, "@see"))
			if ref != "" {
				fd.SeeAlso = append(fd.SeeAlso, ref)
			}

		case strings.HasPrefix(trimmed, "@internal"):
			fd.Internal = true

		case strings.HasPrefix(trimmed, "@set"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@set"))
			name, desc, _ := strings.Cut(rest, " ")
			fd.Set = append(fd.Set, Variable{
				Name:        name,
				Description: strings.TrimSpace(desc),
			})

		case strings.HasPrefix(trimmed, "@noargs"):
			// No-op marker, just indicates the function takes no arguments
		}
	}

	// Finalize
	if inExample && len(exampleLines) > 0 {
		fd.Examples = append(fd.Examples, strings.Join(exampleLines, "\n"))
	}
	if len(descLines) > 0 {
		fd.Description = strings.TrimSpace(strings.Join(descLines, "\n"))
	}

	return fd
}

// parseFileHeader parses file-level doc tags (@file, @brief, @description)
// from a docBlock. Returns the values found.
func parseFileHeader(db docBlock) (name, brief, description string, vars []Variable) {
	if len(db.lines) == 0 {
		return
	}

	var descLines []string
	inDescription := false

	for _, rawLine := range db.lines {
		trimmed := strings.TrimSpace(rawLine)

		if inDescription {
			if isDocTag(trimmed) {
				inDescription = false
			} else if isToolDirective(trimmed) {
				continue // skip tool directives
			} else {
				descLines = append(descLines, rawLine)
				continue
			}
		}

		switch {
		case strings.HasPrefix(trimmed, "@file"):
			name = strings.TrimSpace(strings.TrimPrefix(trimmed, "@file"))
		case strings.HasPrefix(trimmed, "@brief"):
			brief = strings.TrimSpace(strings.TrimPrefix(trimmed, "@brief"))
		case strings.HasPrefix(trimmed, "@description"):
			inDescription = true
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@description"))
			if rest != "" {
				descLines = append(descLines, rest)
			}
		case strings.HasPrefix(trimmed, "@set"):
			rest := strings.TrimSpace(strings.TrimPrefix(trimmed, "@set"))
			n, desc, _ := strings.Cut(rest, " ")
			vars = append(vars, Variable{
				Name:        n,
				Description: strings.TrimSpace(desc),
			})
		}
	}

	if len(descLines) > 0 {
		description = strings.TrimSpace(strings.Join(descLines, "\n"))
	}
	return
}
