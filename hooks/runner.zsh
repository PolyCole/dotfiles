#!/usr/bin/env zsh
# hooks/runner.zsh — machine-composable git hook dispatcher
#
# Usage: hooks/runner.zsh <hook-type> [hook-args...]
#
# Reads the active machine's hooks.conf to determine which scripts to run
# for the given hook type, then executes them in order. Any non-zero exit
# code short-circuits the chain and propagates to git.
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

if [[ ! -f "$hooks_conf" ]]; then
    # No hooks.conf for this machine — nothing to run
    exit 0
fi

# Read hooks.conf and run scripts matching this hook type
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

    sh "$script" "$@"
    exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        echo "Hook $conf_script ($hook_type) failed with exit code $exit_code" >&2
        exit $exit_code
    fi

done < "$hooks_conf"

exit 0
