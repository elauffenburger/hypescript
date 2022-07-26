#!/usr/bin/env bash

# Read the src from stdin.
SRC=$(</dev/stdin)

# Build it!
EXE=$(echo "$SRC" | ./bin/build.sh)
if [[ $? -ne 0 ]]; then
    exit
fi

# Run it!
"$EXE"