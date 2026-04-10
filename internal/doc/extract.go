package doc

import (
	"log/slog"
	"path/filepath"

	"github.com/Hayao0819/seira/internal/depgraph"
	"github.com/Hayao0819/seira/internal/shellparse"
	"mvdan.cc/sh/v3/syntax"
)

// ExtractFile extracts documentation from a single parsed script.
func ExtractFile(script *shellparse.Script, baseDir string) FileDoc {
	relPath, err := filepath.Rel(baseDir, script.FullPath)
	if err != nil {
		relPath = script.FullPath
	}

	fd := FileDoc{
		Path: relPath,
	}

	stmts := script.File.Stmts

	// Extract file-level header from leading comments before any statement.
	if headerDB := extractFileHeader(script.File); len(headerDB.lines) > 0 {
		fd.Name, fd.Brief, fd.Description, fd.Variables = parseFileHeader(headerDB)
	}

	// Track current section
	var currentSection *Section

	for _, stmt := range stmts {
		// Check for @section in comments
		for _, c := range stmt.Comments {
			text := c.Text
			if len(text) > 0 && text[0] == ' ' {
				text = text[1:]
			}
			trimmed := trimLeadingSpace(text)
			if hasPrefix(trimmed, "@section") {
				sectionName := trimPrefix(trimmed, "@section")
				// Save previous section if it has functions
				if currentSection != nil && len(currentSection.Functions) > 0 {
					fd.Sections = append(fd.Sections, *currentSection)
				}
				currentSection = &Section{Name: sectionName}
			}
		}

		fn, ok := stmt.Cmd.(*syntax.FuncDecl)
		if !ok {
			continue
		}

		db := collectDocBlock(stmt.Comments)
		funcDoc := parseFuncDoc(db, fn.Name.Value, int(stmt.Pos().Line()))

		if currentSection != nil {
			currentSection.Functions = append(currentSection.Functions, funcDoc)
		} else {
			fd.Functions = append(fd.Functions, funcDoc)
		}
	}

	// Save last section
	if currentSection != nil && len(currentSection.Functions) > 0 {
		fd.Sections = append(fd.Sections, *currentSection)
	}

	return fd
}

// ExtractProject extracts documentation from all files in a dependency graph.
// excludeDirs specifies directory prefixes (absolute paths) whose files should
// be excluded from documentation (e.g. external package directories).
func ExtractProject(graph *depgraph.Graph, baseDir string, title string, excludeDirs []string) ProjectDoc {
	order, err := graph.TopologicalSort()
	if err != nil {
		slog.Warn("could not topologically sort files, using unordered", "error", err)
		// Fall back to unordered
		for path := range graph.Nodes() {
			order = append(order, path)
		}
	}

	pd := ProjectDoc{
		Title: title,
	}

	for _, path := range order {
		if isExcluded(path, excludeDirs) {
			slog.Debug("excluding external package from documentation", "path", path)
			continue
		}
		node := graph.Node(path)
		if node == nil || node.Script == nil {
			continue
		}
		fd := ExtractFile(node.Script, baseDir)
		// Only include files that have some documentation
		if fd.Name != "" || fd.Brief != "" || fd.Description != "" ||
			len(fd.Functions) > 0 || len(fd.Sections) > 0 {
			pd.Files = append(pd.Files, fd)
		}
	}

	return pd
}

// isExcluded returns true if path is under any of the excluded directories.
func isExcluded(path string, excludeDirs []string) bool {
	for _, dir := range excludeDirs {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			continue
		}
		// If rel doesn't start with "..", path is inside dir
		if !filepath.IsAbs(rel) && (rel == "." || len(rel) < 2 || rel[:2] != "..") {
			return true
		}
	}
	return false
}

// extractFileHeader gathers comments that appear before the first statement
// in the file (excluding shebang lines).
func extractFileHeader(file *syntax.File) docBlock {
	// Collect comments that are before the first statement.
	var firstStmtLine uint
	if len(file.Stmts) > 0 {
		firstStmtLine = file.Stmts[0].Pos().Line()
	}

	var headerComments []syntax.Comment
	for _, stmt := range file.Stmts {
		for _, c := range stmt.Comments {
			if firstStmtLine == 0 || c.Hash.Line() < firstStmtLine {
				headerComments = append(headerComments, c)
			}
		}
	}

	// Also check the file's Last comments (trailing comments not attached to any stmt)
	for _, c := range file.Last {
		if firstStmtLine == 0 || c.Hash.Line() < firstStmtLine {
			headerComments = append(headerComments, c)
		}
	}

	return collectDocBlock(headerComments)
}

func trimLeadingSpace(s string) string {
	for i, c := range s {
		if c != ' ' && c != '\t' {
			return s[i:]
		}
	}
	return ""
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func trimPrefix(s, prefix string) string {
	rest := s[len(prefix):]
	return trimLeadingSpace(rest)
}
