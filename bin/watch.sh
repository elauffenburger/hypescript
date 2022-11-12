#!/usr/bin/env bash

set -o errexit -o pipefail -o nounset

cmd="${@:--x test}"

ignored=$(realpath "${BASH_SOURCE[0]%/*}/../tests/snapshots")
RUST_BACKTRACE=1 cargo watch "$cmd" -c -i "$ignored"