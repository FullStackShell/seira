#!/usr/bin/env bash

source client.sh

# cmd_timeline fetches the home timeline.
# Usage: cmd_timeline [-n limit]
cmd_timeline() {
    local limit=10

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -n|--limit)
                limit="$2"
                shift 2
                ;;
            *)
                echo "Usage: $(basename "$0") timeline [-n limit]" >&2
                return 1
                ;;
        esac
    done

    local body
    body=$(jq -n --argjson limit "$limit" '{limit: $limit}')

    local resp
    resp=$(api_call "notes/timeline" "$body")

    printf '%s' "$resp" | jq -r '.[] | "\(.user.username) [\(.createdAt)] \(.text // "(renote/media)")"'
}

# cmd_local fetches the local timeline.
# Usage: cmd_local [-n limit]
cmd_local() {
    local limit=10

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -n|--limit)
                limit="$2"
                shift 2
                ;;
            *)
                echo "Usage: $(basename "$0") local [-n limit]" >&2
                return 1
                ;;
        esac
    done

    local body
    body=$(jq -n --argjson limit "$limit" '{limit: $limit}')

    local resp
    resp=$(api_call "notes/local-timeline" "$body")

    printf '%s' "$resp" | jq -r '.[] | "\(.user.username) [\(.createdAt)] \(.text // "(renote/media)")"'
}
