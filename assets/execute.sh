#!/bin/sh
# shellcheck disable=all
set -eu
_TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$_TMP_DIR"' EXIT
echo "{{ .TarBallBase64 }}" | base64 -d >"$_TMP_DIR/tarball"
mkdir -p "$_TMP_DIR/extract"
tar -xf "$_TMP_DIR/tarball" -C "$_TMP_DIR/extract"
_EXEC="$_TMP_DIR/execute"
echo '#!{{ .Shebang }}' >"$_EXEC"
echo "export SEIRA_ROOTDIR=\"$_TMP_DIR/extract\"" >>"$_EXEC"
{{ if .HasDeps }}echo "export SEIRA_DEPS_DIR=\"$_TMP_DIR/extract/deps\"" >>"$_EXEC"
{{ end }}echo "cd \"$_TMP_DIR/extract\"" >>"$_EXEC"
{{ if .HasDeps }}cat >>"$_EXEC" <<'SEIRA_PATH_EOF'
{{ .SeiraPathFunc }}
SEIRA_PATH_EOF
{{ end }}echo "source \"$_TMP_DIR/extract/{{ .Entrypoint }}\"" >>"$_EXEC"
echo 'main "$@"' >>"$_EXEC"
chmod +x "$_EXEC"
"$_EXEC" "$@"
