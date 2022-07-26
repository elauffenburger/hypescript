#!/usr/bin/env bash

set -o errexit -o pipefail -o nounset

ignored=$(realpath "${BASH_SOURCE[0]%/*}/../tests/snapshots")
cargo watch -x test -i $ignored