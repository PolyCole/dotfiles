# ************************************
#          Cole's Dotfiles
# ************************************

export DOTFILES="$HOME/repos/dotfiles"

if [[ ! -d "$DOTFILES" ]]; then
  echo "Dotfile directory not found: $DOTFILES"
  return 1
fi

source "$DOTFILES/shell/init.zsh"
source "$DOTFILES/shell/path.zsh"

# Auto-load global modules
for _mod in "$DOTFILES/modules/"*.zsh(N); do
  source "$_mod"
done

# Load machine-specific config
if [[ -n "$DOTFILES_MACHINE" && -d "$DOTFILES/machines/$DOTFILES_MACHINE" ]]; then
  local _mdir="$DOTFILES/machines/$DOTFILES_MACHINE"
  [[ -f "$_mdir/init.zsh"    ]] && source "$_mdir/init.zsh"
  [[ -f "$_mdir/path.zsh"    ]] && source "$_mdir/path.zsh"
  [[ -f "$_mdir/aliases.zsh" ]] && source "$_mdir/aliases.zsh"
  for _mod in "$_mdir/modules/"*.zsh(N); do
    source "$_mod"
  done
fi

unset _mod _mdir

# ls colors
LS_COLORS=$LS_COLORS:'di=1;32:ex=4;31'
export LS_COLORS

# Load discovery system last
[[ -f "$DOTFILES/shell/dots.zsh" ]] && source "$DOTFILES/shell/dots.zsh"
