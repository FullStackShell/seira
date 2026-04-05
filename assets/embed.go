package assets

import _ "embed"

//go:embed execute.sh
var ExecuteTemplate string

//go:embed seira_path.sh
var SeiraPathFunc string
