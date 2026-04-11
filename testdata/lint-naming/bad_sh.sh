#!/bin/sh
# Badly-named POSIX sh library

ANYSOCK_VERSION="0.1.0"
_WRONG_VAR=1

wrong_function() {
    echo "no prefix"
}

anysock_good() {
    echo "this is fine"
}
