#!/usr/bin/env sh
# trufflehog pre-commit hook
# Scans staged changes for secrets before allowing a commit.

if ! command -v trufflehog >/dev/null 2>&1; then
    echo "Warning: trufflehog not found in PATH, skipping secret scan" >&2
    exit 0
fi

# --branch HEAD: without it, --since-commit also walks other local branches'
# history, so old findings elsewhere in the repo block every commit.
# Staged changes are still scanned (reported as "Commit: Staged").
# Unset GIT_INDEX_FILE: in a worktree git passes hooks an absolute index path,
# and trufflehog's internal clone would inherit it and overwrite the real
# index with HEAD, unstaging everything and producing an empty commit.
env -u GIT_INDEX_FILE trufflehog git file://. --since-commit HEAD --branch HEAD --fail
