#!/bin/sh
set -eu
_TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$_TMP_DIR"' EXIT
echo "{{ .TarBallBase64 }}" | base64 -d >"$_TMP_DIR/tarball"
mkdir -p "$_TMP_DIR/extract"
tar -xf "$_TMP_DIR/tarball" -C "$_TMP_DIR/extract"
_EXEC="$_TMP_DIR/execute"
echo '#!{{ .Shebang }}' >"$_EXEC"
echo "export SEIRA_ROOTDIR=\"$_TMP_DIR/extract\"" >>"$_EXEC"
echo "cd \"$_TMP_DIR/extract\"" >>"$_EXEC"
echo "source \"$_TMP_DIR/extract/{{ .Entrypoint }}\"" >>"$_EXEC"
echo 'main "$@"' >>"$_EXEC"
chmod +x "$_EXEC"
"$_EXEC" "$@"
