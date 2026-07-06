#!/usr/bin/env zsh
# hooks/runner.zsh — machine-composable git hook dispatcher
#
# Usage: hooks/runner.zsh <hook-type> [hook-args...]
#
# Reads the active machine's hooks.conf to determine which scripts to run
# for the given hook type, then executes them in order. Any non-zero exit
# code short-circuits the chain and propagates to git.
#
# After the configured scripts, the current repository's own hook
# (.git/hooks/<hook-type>) runs if present — a global core.hooksPath
# REPLACES local hooks, so chaining here keeps per-repo hooks
# (husky, lefthook, hand-written) working.
#
# Environment:
#   DOTFILES         — path to this dotfiles repo (required)
#   DOTFILES_MACHINE — active machine profile (personal, ibotta, unknown)

set -euo pipefail

hook_type="${1:?hook-type required}"
shift

dotfiles="${DOTFILES:-$HOME/repos/dotfiles}"
machine="${DOTFILES_MACHINE:-unknown}"

hooks_conf="$dotfiles/machines/$machine/hooks.conf"

# ---------------------------------------------------------------------------
# Configured machine hooks
# ---------------------------------------------------------------------------
if [[ -f "$hooks_conf" ]]; then
    while IFS= read -r line; do
        # Skip blank lines and comments
        [[ -z "$line" || "$line" == \#* ]] && continue

        conf_type="${line%% *}"
        conf_script="${line##* }"

        [[ "$conf_type" != "$hook_type" ]] && continue

        script="$dotfiles/hooks/$hook_type/$conf_script.sh"

        if [[ ! -f "$script" ]]; then
            echo "Warning: hook script not found: $script" >&2
            continue
        fi

        # '|| exit_code=$?' keeps set -e from aborting before we can report the failure
        exit_code=0
        sh "$script" "$@" || exit_code=$?

        if [[ $exit_code -ne 0 ]]; then
            echo "Hook $conf_script ($hook_type) failed with exit code $exit_code" >&2
            exit $exit_code
        fi

    done < "$hooks_conf"
fi

# ---------------------------------------------------------------------------
# Repository-local hook, if the repo has one
# ---------------------------------------------------------------------------
# git rev-parse --git-path would resolve through core.hooksPath back to us,
# so build the path from --git-dir instead.
git_dir=$(git rev-parse --git-dir 2>/dev/null) || git_dir=""

if [[ -n "$git_dir" && -x "$git_dir/hooks/$hook_type" ]]; then
    exit_code=0
    "$git_dir/hooks/$hook_type" "$@" || exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        echo "Repo-local hook $git_dir/hooks/$hook_type failed with exit code $exit_code" >&2
        exit $exit_code
    fi
fi

exit 0
