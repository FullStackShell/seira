package shellparse

import "mvdan.cc/sh/v3/syntax"

// SourceRef represents a source/. command found in a script.
type SourceRef struct {
	Raw        string     // the original expression as written
	Resolved   string     // after variable default-value expansion
	IsDynamic  bool       // true if path could not be fully resolved statically
	Directives Directives // seira directives from preceding comments
}

// Script holds parsed metadata about a shell script.
type Script struct {
	Sources       []SourceRef
	TopLevelFuncs []syntax.FuncDecl
	File          *syntax.File
	FullPath      string // absolute path to the script
}

// HasFunc returns true if the script declares a top-level function with the given name.
func (s *Script) HasFunc(name string) bool {
	for _, fn := range s.TopLevelFuncs {
		if fn.Name.Value == name {
			return true
		}
	}
	return false
}
