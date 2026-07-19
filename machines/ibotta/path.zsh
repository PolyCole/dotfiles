# machines/ibotta/path.zsh — Ibotta-specific PATH modifications

# ---------------------------------------------------------------------------
# Core PATH additions
# ---------------------------------------------------------------------------
export PATH="$PATH:$HOME/.local/bin"

# Toolbox scripts
export PATH="$HOME/toolbox/ibotta_scripts:$PATH"
export PATH="$HOME/toolbox/ibotta_scripts/ibotta_linkers:$PATH"
export PATH="$HOME/toolbox/found_scripts:$PATH"

# npm global bin
path+=("$HOME/npm/bin")

# Go install bin (toil CLI and other `go install` binaries)
export PATH="$PATH:$HOME/go/bin"

# ---------------------------------------------------------------------------
# rbenv
# ---------------------------------------------------------------------------
command -v rbenv >/dev/null && eval "$(rbenv init - zsh)"

# ---------------------------------------------------------------------------
# NVM
# ---------------------------------------------------------------------------
# $HOMEBREW_PREFIX is set by 'brew shellenv' in shell/path.zsh
export NVM_DIR="$HOME/.nvm"
[ -s "$HOMEBREW_PREFIX/opt/nvm/nvm.sh" ] && . "$HOMEBREW_PREFIX/opt/nvm/nvm.sh"
[ -s "$HOMEBREW_PREFIX/opt/nvm/etc/bash_completion.d/nvm" ] && \. "$HOMEBREW_PREFIX/opt/nvm/etc/bash_completion.d/nvm"

# ---------------------------------------------------------------------------
# conda
# ---------------------------------------------------------------------------
__conda_setup="$('/Users/cole.polyak/miniconda3/bin/conda' 'shell.zsh' 'hook' 2> /dev/null)"
if [ $? -eq 0 ]; then
    eval "$__conda_setup"
else
    if [ -f "/Users/cole.polyak/miniconda3/etc/profile.d/conda.sh" ]; then
        . "/Users/cole.polyak/miniconda3/etc/profile.d/conda.sh"
    else
        export PATH="/Users/cole.polyak/miniconda3/bin:$PATH"
    fi
fi
unset __conda_setup

# ---------------------------------------------------------------------------
# Antigravity
# ---------------------------------------------------------------------------
export PATH="/Users/cole.polyak/.antigravity/antigravity/bin:$PATH"

# ---------------------------------------------------------------------------
# bun
# ---------------------------------------------------------------------------
[ -s "/Users/cole.polyak/.bun/_bun" ] && source "/Users/cole.polyak/.bun/_bun"
export BUN_INSTALL="$HOME/.bun"
export PATH="$BUN_INSTALL/bin:$PATH"
