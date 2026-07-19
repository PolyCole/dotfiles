#!/usr/bin/env sh
# beads prepare-commit-msg hook
# Delegates to 'bd hooks run prepare-commit-msg' which contains the actual hook logic.

if ! command -v bd >/dev/null 2>&1; then
    echo "Warning: bd command not found in PATH, skipping beads prepare-commit-msg hook" >&2
    exit 0
fi

exec bd hooks run prepare-commit-msg "$@"
