# ************************************
#          Cole's Dotfiles
# ************************************

# Disable p10k instant prompt — startup banner intentionally outputs during init
typeset -g POWERLEVEL9K_INSTANT_PROMPT=off

# Checking for the existance of our dotfile repo.
# Respect a pre-set $DOTFILES so alternate clone locations work.
export DOTFILES="${DOTFILES:-$HOME/repos/dotfiles}"
if [ ! -d "$DOTFILES" ]; then
  RED='\033[0;31m'
  NC='\033[0m'

  echo "
  ${NC}x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x
      ${RED}Dotfile Directory not found!
  ${NC}x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x.x
  "
  # 'exit' here would kill the terminal — this file is sourced.
  return
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

# bun completions
[ -s "/private/tmp/claude-502/-Users-cole-polyak-repos-turbos-upa-weirdness/3f83125b-04c9-4b58-ba83-e33c7fa8d0e3/scratchpad/bun142/_bun" ] && source "/private/tmp/claude-502/-Users-cole-polyak-repos-turbos-upa-weirdness/3f83125b-04c9-4b58-ba83-e33c7fa8d0e3/scratchpad/bun142/_bun"
