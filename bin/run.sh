#!/usr/bin/env bash

# Read the src from stdin.
SRC=$(</dev/stdin)

# Build it!
EXE=$(echo "$SRC" | ./bin/build.sh)

# Run it!
"$EXE"