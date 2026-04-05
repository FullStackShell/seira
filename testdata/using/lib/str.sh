#!/usr/bin/env bash

str::upper() {
    echo "$1" | tr '[:lower:]' '[:upper:]'
}

str::lower() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}
