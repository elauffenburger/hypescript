#!/usr/bin/env bash

set -o errexit -o pipefail -o nounset

# Build hsc.
cargo build --bin hsc

# Make a dir for build output.
rm -rf ./build
mkdir ./build

build_dir=./build
src_dir="$build_dir"/src
bin_dir="$build_dir"/bin

# Forward the script args to hsc.
cargo run --bin hsc -- -o "$build_dir" "$@"

# Format the output files.
clang-format $src_dir/* -i

# Compilation time!

# Make the bin dir
mkdir "$bin_dir"

# Compile the runtime and main obj files.
clang++ -g -c -std=c++20 "$src_dir"/runtime.cpp -o "$bin_dir"/runtime.o
clang++ -g -c -std=c++20 "$src_dir"/main.cpp -o "$bin_dir"/main.o

# Compile the executable.
clang++ -g -std=c++20 "$bin_dir"/*.o -o "$bin_dir"/out

# Run the compiled binary.
"$bin_dir"/out