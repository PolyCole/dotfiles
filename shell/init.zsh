# shell/init.zsh — machine detection, oh-my-zsh, p10k

# ---------------------------------------------------------------------------
# Machine detection
# ---------------------------------------------------------------------------
_hostname=$(hostname -s)

if [[ "$_hostname" == "six" ]]; then
  export DOTFILES_MACHINE="personal"
elif [[ "$_hostname" =~ ^[A-Z]{1}[A-Z0-9-]+$ ]]; then
  export DOTFILES_MACHINE="ibotta"
else
  export DOTFILES_MACHINE="unknown"
fi

unset _hostname

# ---------------------------------------------------------------------------
# Linker aliases (dynamically generated, optional)
# ---------------------------------------------------------------------------
[[ -f "$HOME/.linker_aliases" ]] && source "$HOME/.linker_aliases"

# ---------------------------------------------------------------------------
# oh-my-zsh
# ---------------------------------------------------------------------------
export ZSH="$HOME/.oh-my-zsh"

if [[ "$DOTFILES_MACHINE" == "personal" ]]; then
  source /opt/homebrew/opt/powerlevel10k/powerlevel10k.zsh-theme
else
  ZSH_THEME="powerlevel10k/powerlevel10k"
fi

plugins=(
  git
  zsh-osx-keychain
)

source "$ZSH/oh-my-zsh.sh"

[[ -f "$HOME/.p10k.zsh" ]] && source "$HOME/.p10k.zsh"
