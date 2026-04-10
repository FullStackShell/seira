package doc

import (
	"fmt"
	"html"
	"io"
	"strings"

	"github.com/Hayao0819/seira/internal/doc/assets"
)

// renderHTMLHeader writes the HTML document header using the embedded template.
func renderHTMLHeader(hw *htmlWriter, title string) {
	header := assets.DocHeader
	header = strings.ReplaceAll(header, "{{.Title}}", html.EscapeString(title))
	header = strings.ReplaceAll(header, "{{.CSS}}", assets.DocCSS)
	hw.writef("%s", header)
}

// renderHTMLFooter writes the HTML document footer using the embedded template.
func renderHTMLFooter(hw *htmlWriter) {
	hw.writef("%s", assets.DocFooter)
}

// RenderHTML writes the project documentation in HTML format.
func RenderHTML(w io.Writer, pd ProjectDoc) error {
	hw := &htmlWriter{w: w}

	title := pd.Title
	if title == "" {
		title = "Shell Script Documentation"
	}

	renderHTMLHeader(hw, title)

	// Header
	hw.writef("<header>\n<h1>%s</h1>\n</header>\n", html.EscapeString(title))

	// Sidebar with TOC
	if len(pd.Files) > 1 {
		hw.writef("<nav class=\"sidebar\">\n<h2>Files</h2>\n<ul>\n")
		for _, f := range pd.Files {
			name := f.Name
			if name == "" {
				name = f.Path
			}
			anchor := tocAnchor(name)
			hw.writef("<li><a href=\"#%s\">%s</a>\n", anchor, html.EscapeString(name))
			// Nested function list
			allFuncs := collectVisibleFuncs(f)
			if len(allFuncs) > 0 {
				hw.writef("<ul>\n")
				for _, fn := range allFuncs {
					hw.writef("<li><a href=\"#%s\">%s()</a></li>\n",
						funcAnchor(name, fn.Name), html.EscapeString(fn.Name))
				}
				hw.writef("</ul>\n")
			}
			hw.writef("</li>\n")
		}
		hw.writef("</ul>\n</nav>\n")
	}

	hw.writef("<main>\n")

	for _, f := range pd.Files {
		renderHTMLFile(hw, f)
	}

	renderHTMLFooter(hw)

	return hw.err
}

// RenderHTMLFile writes documentation for a single file.
func RenderHTMLFile(w io.Writer, fd FileDoc) error {
	hw := &htmlWriter{w: w}

	title := fd.Name
	if title == "" {
		title = fd.Path
	}

	renderHTMLHeader(hw, title)
	hw.writef("<main>\n")

	renderHTMLFile(hw, fd)

	renderHTMLFooter(hw)

	return hw.err
}

func renderHTMLFile(hw *htmlWriter, fd FileDoc) {
	name := fd.Name
	if name == "" {
		name = fd.Path
	}
	anchor := tocAnchor(name)

	hw.writef("<section class=\"file\" id=\"%s\">\n", anchor)
	hw.writef("<h2>%s</h2>\n", html.EscapeString(name))

	if fd.Path != "" && fd.Name != "" {
		hw.writef("<p class=\"file-path\"><code>%s</code></p>\n", html.EscapeString(fd.Path))
	}
	if fd.Brief != "" {
		hw.writef("<p class=\"brief\">%s</p>\n", html.EscapeString(fd.Brief))
	}
	if fd.Description != "" {
		hw.writef("<div class=\"description\">%s</div>\n", renderTextToHTML(fd.Description))
	}

	// File-level variables
	if len(fd.Variables) > 0 {
		hw.writef("<h3>Global Variables</h3>\n<dl class=\"variables\">\n")
		for _, v := range fd.Variables {
			hw.writef("<dt><code>%s</code></dt>\n<dd>%s</dd>\n",
				html.EscapeString(v.Name), html.EscapeString(v.Description))
		}
		hw.writef("</dl>\n")
	}

	// Functions not in a section
	for _, fn := range fd.Functions {
		renderHTMLFunc(hw, fn, name)
	}

	// Sections
	for _, sec := range fd.Sections {
		hw.writef("<div class=\"section\">\n<h3>%s</h3>\n", html.EscapeString(sec.Name))
		for _, fn := range sec.Functions {
			renderHTMLFunc(hw, fn, name)
		}
		hw.writef("</div>\n")
	}

	hw.writef("</section>\n")
}

func renderHTMLFunc(hw *htmlWriter, fn FuncDoc, fileName string) {
	if fn.Internal {
		return
	}

	anchor := funcAnchor(fileName, fn.Name)
	hw.writef("<article class=\"function\" id=\"%s\">\n", anchor)
	hw.writef("<h4><code>%s()</code></h4>\n", html.EscapeString(fn.Name))

	if fn.Description != "" {
		hw.writef("<div class=\"description\">%s</div>\n", renderTextToHTML(fn.Description))
	}

	// Parameters
	if len(fn.Params) > 0 {
		hw.writef("<h5>Arguments</h5>\n<table class=\"params\">\n")
		hw.writef("<thead><tr><th>Name</th><th>Description</th></tr></thead>\n<tbody>\n")
		for _, p := range fn.Params {
			hw.writef("<tr><td><code>%s</code></td><td>%s</td></tr>\n",
				html.EscapeString(p.Name), html.EscapeString(p.Description))
		}
		hw.writef("</tbody>\n</table>\n")
	}

	// Options
	if len(fn.Options) > 0 {
		hw.writef("<h5>Options</h5>\n<table class=\"options\">\n")
		hw.writef("<thead><tr><th>Flag</th><th>Description</th></tr></thead>\n<tbody>\n")
		for _, o := range fn.Options {
			hw.writef("<tr><td><code>%s</code></td><td>%s</td></tr>\n",
				html.EscapeString(o.Flags), html.EscapeString(o.Description))
		}
		hw.writef("</tbody>\n</table>\n")
	}

	// Stdin / Stdout
	if fn.Stdin != "" {
		hw.writef("<p><strong>Stdin:</strong> %s</p>\n", html.EscapeString(fn.Stdin))
	}
	if fn.Stdout != "" {
		hw.writef("<p><strong>Stdout:</strong> %s</p>\n", html.EscapeString(fn.Stdout))
	}

	// Return
	if fn.Returns != "" {
		hw.writef("<p><strong>Returns:</strong> %s</p>\n", html.EscapeString(fn.Returns))
	}

	// Exit codes
	if len(fn.ExitCodes) > 0 {
		hw.writef("<h5>Exit Codes</h5>\n<table class=\"exitcodes\">\n")
		hw.writef("<thead><tr><th>Code</th><th>Description</th></tr></thead>\n<tbody>\n")
		for _, ec := range fn.ExitCodes {
			hw.writef("<tr><td><code>%s</code></td><td>%s</td></tr>\n",
				html.EscapeString(ec.Code), html.EscapeString(ec.Description))
		}
		hw.writef("</tbody>\n</table>\n")
	}

	// Set variables
	if len(fn.Set) > 0 {
		hw.writef("<h5>Sets</h5>\n<dl class=\"variables\">\n")
		for _, v := range fn.Set {
			hw.writef("<dt><code>%s</code></dt>\n<dd>%s</dd>\n",
				html.EscapeString(v.Name), html.EscapeString(v.Description))
		}
		hw.writef("</dl>\n")
	}

	// Examples
	for _, ex := range fn.Examples {
		hw.writef("<h5>Example</h5>\n<pre><code>%s</code></pre>\n",
			html.EscapeString(strings.TrimSpace(ex)))
	}

	// See also
	if len(fn.SeeAlso) > 0 {
		hw.writef("<p class=\"see-also\"><strong>See Also:</strong> ")
		for i, ref := range fn.SeeAlso {
			if i > 0 {
				hw.writef(", ")
			}
			hw.writef("<code>%s</code>", html.EscapeString(ref))
		}
		hw.writef("</p>\n")
	}

	hw.writef("</article>\n")
}

// collectVisibleFuncs returns all non-internal functions across sections and top-level.
func collectVisibleFuncs(fd FileDoc) []FuncDoc {
	var result []FuncDoc
	for _, fn := range fd.Functions {
		if !fn.Internal {
			result = append(result, fn)
		}
	}
	for _, sec := range fd.Sections {
		for _, fn := range sec.Functions {
			if !fn.Internal {
				result = append(result, fn)
			}
		}
	}
	return result
}

func funcAnchor(fileName, funcName string) string {
	return tocAnchor(fileName) + "--" + tocAnchor(funcName)
}

// renderTextToHTML converts plain text to simple HTML paragraphs.
func renderTextToHTML(text string) string {
	paragraphs := strings.Split(text, "\n\n")
	var parts []string
	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, "<p>"+html.EscapeString(p)+"</p>")
		}
	}
	return strings.Join(parts, "\n")
}

type htmlWriter struct {
	w   io.Writer
	err error
}

func (hw *htmlWriter) writef(format string, args ...any) {
	if hw.err != nil {
		return
	}
	_, hw.err = fmt.Fprintf(hw.w, format, args...)
}
