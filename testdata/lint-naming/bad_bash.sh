#!/bin/bash
# Badly-named bash library with violations

declare -g _L=""
declare -gA _WRONG_NAME=()

bad_function() {
    echo "no prefix"
}

oop::good() {
    echo "this is fine"
}
