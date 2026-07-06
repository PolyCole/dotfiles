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
