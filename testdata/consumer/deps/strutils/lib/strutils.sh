#!/usr/bin/env bash

str_upper() {
    echo "$1" | tr '[:lower:]' '[:upper:]'
}

str_lower() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}
