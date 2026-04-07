#!/usr/bin/env bash

source client.sh

# cmd_notifications fetches recent notifications.
# Usage: cmd_notifications [-n limit]
cmd_notifications() {
    local limit=10

    while [[ $# -gt 0 ]]; do
        case "$1" in
            -n|--limit)
                limit="$2"
                shift 2
                ;;
            *)
                echo "Usage: $(basename "$0") notifications [-n limit]" >&2
                return 1
                ;;
        esac
    done

    local body
    body=$(jq -n --argjson limit "$limit" '{limit: $limit}')

    local resp
    resp=$(api_call "i/notifications" "$body")

    printf '%s' "$resp" | jq -r '.[] | "\(.type) from \(.user.username // "system") - \(.note.text // .message // "(no text)")"'
}
