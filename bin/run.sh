#!/usr/bin/env bash

usage() {
    echo "usage: $0 [src]"
}

main() {
    src=

    while :; do
        case "$1" in
        '')
            break
            ;;

        *)
            # Read the source from the provided path.
            if [[ -z "$src" ]]; then
                src=$(cat "$1")
                shift

                continue
            fi

            # If we've already read src, this is an unknown flag.
            usage
            exit 1
            ;;
        esac
    done

    # Read the src from stdin.
    if [[ -z "$src" ]]; then
        src=$(cat /dev/stdin)
    fi

    # Build it!
    exe=$(echo "$src" | ./bin/build.sh)
    if [[ $? -ne 0 ]]; then
        exit
    fi

    # Run it!
    "$exe"
}

main "$@"
