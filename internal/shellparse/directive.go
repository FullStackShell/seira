package shellparse

import (
	"strings"

	"mvdan.cc/sh/v3/syntax"
)

// directivePrefix is matched against Comment.Text which does NOT include the '#'.
// For "# @seira:ignore", Comment.Text is " @seira:ignore".
const directivePrefix = " @seira:"

// Directive represents a parsed seira directive from a comment.
type Directive struct {
	Name string     // e.g., "ignore", "type", "param"
	Args string     // raw argument string after the directive name
	Pos  syntax.Pos // source position of the comment
}

// Directives is a set of directives attached to a statement.
type Directives []Directive

// Has returns true if a directive with the given name exists.
func (d Directives) Has(name string) bool {
	for _, dir := range d {
		if dir.Name == name {
			return true
		}
	}
	return false
}

// Get returns the first directive with the given name.
func (d Directives) Get(name string) (Directive, bool) {
	for _, dir := range d {
		if dir.Name == name {
			return dir, true
		}
	}
	return Directive{}, false
}

// GetAll returns all directives with the given name.
func (d Directives) GetAll(name string) []Directive {
	var result []Directive
	for _, dir := range d {
		if dir.Name == name {
			result = append(result, dir)
		}
	}
	return result
}

// ParseDirectives extracts seira directives from statement comments.
// Comments matching "# @seira:<name> [args]" are parsed into Directive values.
// Non-matching comments are silently ignored.
func ParseDirectives(comments []syntax.Comment) Directives {
	var directives Directives
	for _, c := range comments {
		text := c.Text
		if !strings.HasPrefix(text, directivePrefix) {
			continue
		}
		body := strings.TrimPrefix(text, directivePrefix)
		body = strings.TrimSpace(body)
		if body == "" {
			continue
		}

		name, args, _ := strings.Cut(body, " ")
		directives = append(directives, Directive{
			Name: name,
			Args: strings.TrimSpace(args),
			Pos:  c.Hash,
		})
	}
	return directives
}
