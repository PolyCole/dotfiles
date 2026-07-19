#!/usr/bin/env sh
# beads post-checkout hook
# Delegates to 'bd hook post-checkout' which contains the actual hook logic.

if ! command -v bd >/dev/null 2>&1; then
    exit 0
fi

exec bd hooks run post-checkout "$@"
