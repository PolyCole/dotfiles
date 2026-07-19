#!/usr/bin/env sh
# beads post-merge hook
# Delegates to 'bd hook post-merge' which contains the actual hook logic.

if ! command -v bd >/dev/null 2>&1; then
    echo "Warning: bd command not found in PATH, skipping beads post-merge hook" >&2
    exit 0
fi

exec bd hooks run post-merge "$@"
