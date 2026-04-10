package assets

import _ "embed"

//go:embed doc.css
var DocCSS string

//go:embed doc_header.html
var DocHeader string

//go:embed doc_footer.html
var DocFooter string
