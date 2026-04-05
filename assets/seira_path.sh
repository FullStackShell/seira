
# shellcheck disable=2148
# seira_path: resolve a package file path
# Usage: seira_path owner/repo/path/to/file
# Returns the absolute path to the file in the deps directory.
#
# SEIRA_DEPS_DIR can be set to override the deps directory location.
# Default: ./deps (relative to project root or extracted archive)
seira_path() {
    local _ref="$1"
    local _deps_dir="${SEIRA_DEPS_DIR:-./deps}"

    # Parse: owner/repo/path → repo/path
    local _without_owner="${_ref#*/}"

    if [ -z "$_without_owner" ] || [ "$_without_owner" = "$_ref" ]; then
        echo "seira_path: invalid reference: $_ref (expected owner/repo[/path])" >&2
        return 1
    fi

    local _resolved="$_deps_dir/$_without_owner"

    if [ ! -e "$_resolved" ]; then
        echo "seira_path: not found: $_resolved" >&2
        return 1
    fi

    # Return absolute path
    cd -P "$(dirname "$_resolved")" 2>/dev/null && echo "$(pwd)/$(basename "$_resolved")"
}
