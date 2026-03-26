#!/usr/bin/env sh
# beads pre-commit hook
# Delegates to 'bd hook pre-commit' which contains the actual hook logic.

if ! command -v bd >/dev/null 2>&1; then
    echo "Warning: bd command not found in PATH, skipping beads pre-commit hook" >&2
    exit 0
fi

exec bd hooks run pre-commit "$@"
