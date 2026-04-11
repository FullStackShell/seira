#!/bin/bash
# Well-named bash library following naming conventions

declare -gA _SEIRA_OOP=()
declare -gi _SEIRA_OOP_NEXT_ID=0

oop::class() {
    echo "class"
}

oop::_internal() {
    echo "internal"
}

# OOP class method - should be exempt
Animal::speak() {
    echo "meow"
}
