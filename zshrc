# ************************************
#          Cole's Dotfiles
# ************************************

# Disable p10k instant prompt — startup banner intentionally outputs during init
typeset -g POWERLEVEL9K_INSTANT_PROMPT=off

# Checking for the existance of our dotfile repo.
export DOTFILES="$HOME/repos/dotfiles"
if [ ! -d "$DOTFILES" ]; then
  RED='\033[0;31m'
  NC='\033[0m'

  echo "
  ${NC}x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x
      ${RED}Dotfile Directory not found!
  ${NC}x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x
  "
  exit
fi;

# Core shell bootstrap: machine detection, oh-my-zsh, p10k
source $DOTFILES/shell/init.zsh

# Universal PATH modifications
source $DOTFILES/shell/path.zsh

# dots discovery system
source $DOTFILES/shell/dots.zsh

# Shared modules (available on all machines)
for _m in $DOTFILES/modules/*.zsh(N); do
  source "$_m"
done
unset _m

# Machine-specific config
if [[ -n "$DOTFILES_MACHINE" ]]; then
  [[ -f "$DOTFILES/machines/$DOTFILES_MACHINE/path.zsh" ]]    && source "$DOTFILES/machines/$DOTFILES_MACHINE/path.zsh"
  [[ -f "$DOTFILES/machines/$DOTFILES_MACHINE/init.zsh" ]]    && source "$DOTFILES/machines/$DOTFILES_MACHINE/init.zsh"
  [[ -f "$DOTFILES/machines/$DOTFILES_MACHINE/aliases.zsh" ]] && source "$DOTFILES/machines/$DOTFILES_MACHINE/aliases.zsh"
  for _m in $DOTFILES/machines/$DOTFILES_MACHINE/modules/*.zsh(N); do
    source "$_m"
  done
  unset _m
fi

# ls colors
LS_COLORS=$LS_COLORS:'di=1;32:ex=4;31' ; export LS_COLORS
