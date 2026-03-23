# machines/ibotta/path.zsh — Ibotta-specific PATH modifications

# ---------------------------------------------------------------------------
# Core PATH additions
# ---------------------------------------------------------------------------
export PATH="$PATH:$HOMEBREW_PREFIX/opt/coreutils/libexec/gnubin"
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
if which rbenv > /dev/null; then eval "$(rbenv init - zsh)"; fi

# ---------------------------------------------------------------------------
# NVM
# ---------------------------------------------------------------------------
export NVM_DIR="$HOME/.nvm"
[ -s "$(brew --prefix)/opt/nvm/nvm.sh" ] && . "$(brew --prefix)/opt/nvm/nvm.sh"
[ -s "$(brew --prefix)/opt/nvm/etc/bash_completion.d/nvm" ] && \. "$(brew --prefix)/opt/nvm/etc/bash_completion.d/nvm"
