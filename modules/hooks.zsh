# modules/hooks.zsh
# Git hooks management via the DOTS composable hooks system

# Commands:
#   dots-hooks-install    - Install git hooks globally using core.hooksPath

# Installs the DOTS git hook system globally.
# Sets git's core.hooksPath to ~/.git-hooks and writes thin dispatcher
# scripts there for each supported hook type. Each dispatcher calls
# hooks/runner.zsh with the appropriate hook type.
dots-hooks-install() {
    local dotfiles="${DOTFILES:-$HOME/repos/dotfiles}"
    local hooks_dir="$HOME/.git-hooks"
    local runner="$dotfiles/hooks/runner.zsh"

    if [[ ! -f "$runner" ]]; then
        echo "Error: runner not found at $runner" >&2
        return 1
    fi

    mkdir -p "$hooks_dir"

    local hook_types=(
        pre-commit
        post-checkout
        post-merge
        pre-push
        prepare-commit-msg
    )

    for hook_type in "${hook_types[@]}"; do
        local hook_file="$hooks_dir/$hook_type"
        cat > "$hook_file" <<EOF
#!/usr/bin/env sh
# DOTS hook dispatcher — $hook_type
exec "$runner" "$hook_type" "\$@"
EOF
        chmod +x "$hook_file"
        echo "  wrote $hook_file"
    done

    git config --global core.hooksPath "$hooks_dir"
    echo "Installed DOTS hooks to $hooks_dir (core.hooksPath set globally)"
}
