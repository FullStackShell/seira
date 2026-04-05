package bundler

import (
	"encoding/base64"
	"io"
	"text/template"

	"github.com/Hayao0819/seira/assets"
)

type templateData struct {
	TarBallBase64 string
	Shebang       string
	Entrypoint    string
	HasDeps       bool
	SeiraPathFunc string
}

// renderOutput renders the execute.sh template with the tarball data.
func renderOutput(w io.Writer, tarball []byte, shebang string, entrypoint string, hasDeps bool) error {
	tmpl, err := template.New("execute").Parse(assets.ExecuteTemplate)
	if err != nil {
		return err
	}

	data := templateData{
		TarBallBase64: base64.StdEncoding.EncodeToString(tarball),
		Shebang:       shebang,
		Entrypoint:    entrypoint,
		HasDeps:       hasDeps,
		SeiraPathFunc: assets.SeiraPathFunc,
	}

	return tmpl.Execute(w, data)
}
