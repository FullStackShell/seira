#!/usr/bin/env bash

source client.sh

# cmd_search searches notes by keyword.
# Usage: cmd_search [-n limit] <query>
cmd_search() {
    local limit=10

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -n|--limit)
                limit="$2"
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

    local query="$*"
    if [[ -z "$query" ]]; then
        echo "Usage: $(basename "$0") search [-n limit] <query>" >&2
        return 1
    fi

    local body
    body=$(jq -n --arg q "$query" --argjson limit "$limit" \
        '{query: $q, limit: $limit}')

    local resp
    resp=$(api_call "notes/search" "$body")

    printf '%s' "$resp" | jq -r '.[] | "\(.user.username) [\(.createdAt)] \(.text // "(renote/media)")"'
}
