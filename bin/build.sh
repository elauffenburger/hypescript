#!/usr/bin/env bash

set -o errexit -o pipefail -o nounset

# Read the src from stdin.
SRC=$(</dev/stdin)

# Build hsc.
cargo build --bin hsc

# Make a dir for build output.
rm -rf ./build
mkdir ./build
BUILD_DIR=./build
SRC_DIR="$BUILD_DIR"/src
BIN_DIR="$BUILD_DIR"/bin

# Pipe src to hsc and compile to the temp dir.
echo "$SRC" | ./target/debug/hsc -o "$BUILD_DIR"

# Format the output files.
clang-format $SRC_DIR/* -i

# Compilation time!

# Make the bin dir
mkdir "$BIN_DIR"

# Compile the runtime and main obj files.
clang++ -g -c -std=c++20 "$SRC_DIR"/runtime.cpp -o "$BIN_DIR"/runtime.o
clang++ -g -c -std=c++20 "$SRC_DIR"/main.cpp -o "$BIN_DIR"/main.o

# Compile the executable.
clang++ -g -std=c++20 "$BIN_DIR"/*.o -o "$BIN_DIR"/out

# Return the location of the compiled binary.
echo "$BIN_DIR"/out