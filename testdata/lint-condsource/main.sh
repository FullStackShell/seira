#!/usr/bin/env bash

if [[ -f lib/optional.sh ]]; then
    source lib/optional.sh
fi

source lib/helper.sh

main() {
    greet
}
