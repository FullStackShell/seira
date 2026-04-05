#!/usr/bin/env bash

#shellcheck source=/dev/null
source hoge-2.sh

sayHello() {
    echo "Hello World!"
}

nestedFunction() {
    hoge() {
        echo "Hoge"
    }
}
