package shellparse

import (
	"io"
	"os"
	"path/filepath"

	"mvdan.cc/sh/v3/syntax"
)

// Parser wraps syntax.Parser without global mutable state.
type Parser struct {
	p *syntax.Parser
}

func NewParser(opts ...syntax.ParserOption) *Parser {
	return &Parser{p: syntax.NewParser(opts...)}
}

func (p *Parser) Parse(r io.Reader, name string) (*syntax.File, error) {
	return p.p.Parse(r, name)
}

func (p *Parser) ParseFile(path string) (*syntax.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Parse(f, path)
}

// Analyze parses a script and extracts all metadata (sources, functions).
func (p *Parser) Analyze(r io.Reader, name string) (*Script, error) {
	parsed, err := p.Parse(r, name)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}

	sources := extractSources(parsed)
	funcs := extractTopLevelFuncs(parsed.Stmts)

	return &Script{
		Sources:       sources,
		TopLevelFuncs: funcs,
		File:          parsed,
		FullPath:      absPath,
	}, nil
}

// AnalyzeFile is a convenience method that opens, parses, and analyzes a file.
func (p *Parser) AnalyzeFile(path string) (*Script, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return p.Analyze(f, path)
}
