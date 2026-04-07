#!/usr/bin/env bash

source client.sh

# cmd_post creates a new note.
# Usage: cmd_post [-v public|home|followers|specified] [text...]
cmd_post() {
    local visibility="public"

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -v|--visibility)
                visibility="$2"
                shift 2
                ;;
            --)
                shift
                break
                ;;
            *)
                break
                ;;
        esac
    done

    local text="$*"
    if [[ -z "$text" ]]; then
        echo "Usage: $(basename "$0") post [-v visibility] <text>" >&2
        return 1
    fi

    local body
    body=$(jq -n --arg text "$text" --arg vis "$visibility" \
        '{text: $text, visibility: $vis}')

    local resp
    resp=$(api_call "notes/create" "$body")

    local note_id
    note_id=$(printf '%s' "$resp" | jq -r '.createdNote.id // empty')
    if [[ -n "$note_id" ]]; then
        echo "Posted: $note_id"
    else
        echo "Error: $(printf '%s' "$resp" | jq -r '.error.message // "unknown error"')" >&2
        return 1
    fi
}
