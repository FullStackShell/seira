#!/usr/bin/env bash

source lib/client.sh
source lib/post.sh
source lib/timeline.sh
source lib/search.sh
source lib/notifications.sh

usage() {
    cat <<'USAGE'
misskey-client - A simple Misskey CLI client

Environment:
  MISSKEY_HOST   Instance URL (e.g. https://misskey.io)
  MISSKEY_TOKEN  API access token

Commands:
  post [-v visibility] <text>   Create a new note
  timeline [-n limit]           Show home timeline
  local [-n limit]              Show local timeline
  search [-n limit] <query>     Search notes
  notifications [-n limit]      Show notifications
  help                          Show this help

Examples:
  misskey-client post "Hello, Fediverse!"
  misskey-client post -v home "Home only post"
  misskey-client timeline -n 5
  misskey-client search "seira"
USAGE
}

main() {
    check_deps || return 1

    local cmd="${1:-help}"
    shift 2>/dev/null || true

    case "$cmd" in
        post)
            check_config || return 1
            cmd_post "$@"
            ;;
        timeline|tl)
            check_config || return 1
            cmd_timeline "$@"
            ;;
        local|ltl)
            check_config || return 1
            cmd_local "$@"
            ;;
        search)
            check_config || return 1
            cmd_search "$@"
            ;;
        notifications|notif)
            check_config || return 1
            cmd_notifications "$@"
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            echo "Unknown command: $cmd" >&2
            usage >&2
            return 1
            ;;
    esac
}
