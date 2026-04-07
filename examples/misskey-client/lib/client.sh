#!/usr/bin/env bash

# Misskey API client base
# All Misskey API endpoints are POST with JSON body.

MISSKEY_HOST="${MISSKEY_HOST:-}"
MISSKEY_TOKEN="${MISSKEY_TOKEN:-}"

# api_call <endpoint> [json_body]
# Calls a Misskey API endpoint and returns the JSON response.
api_call() {
    local endpoint="$1"
    local body="${2:-{\}}"
    local url="${MISSKEY_HOST}/api/${endpoint}"

    # Inject token into the body
    if [[ -n "$MISSKEY_TOKEN" ]]; then
        body=$(printf '%s' "$body" | jq --arg token "$MISSKEY_TOKEN" '. + {i: $token}')
    fi

    curl -s -X POST "$url" \
        -H "Content-Type: application/json" \
        -d "$body"
}

# check_deps ensures required commands are available.
check_deps() {
    local missing=()
    for cmd in curl jq; do
        if ! command -v "$cmd" &>/dev/null; then
            missing+=("$cmd")
        fi
    done
    if [[ ${#missing[@]} -gt 0 ]]; then
        echo "Error: missing required commands: ${missing[*]}" >&2
        return 1
    fi
}

# check_config ensures MISSKEY_HOST and MISSKEY_TOKEN are set.
check_config() {
    if [[ -z "$MISSKEY_HOST" ]]; then
        echo "Error: MISSKEY_HOST is not set" >&2
        echo "  export MISSKEY_HOST=https://misskey.example.com" >&2
        return 1
    fi
    if [[ -z "$MISSKEY_TOKEN" ]]; then
        echo "Error: MISSKEY_TOKEN is not set" >&2
        echo "  export MISSKEY_TOKEN=your_api_token" >&2
        return 1
    fi
}
