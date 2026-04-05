#!/usr/bin/env bash

source lib/helper.sh

SCRIPT_DIR="${BASH_SOURCE[0]%/*}"

main() {
    echo "hello from $SCRIPT_DIR"
}
