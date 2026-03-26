# ************************************
#          Cole's Dotfiles
# ************************************

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

# Begin sourcing things.
source $DOTFILES/.zsh_config
source $DOTFILES/common/.common_config


# We'll decide on which specific configuration we want to run
# based on the machine we're on. Things that I want to be common across
# all my machines can simply be placed outside of this block.
if (( $(hostname -s) != six )); then
  source $DOTFILES/work/.work_config
else
  source $DOTFILES/personal/.personal_config
fi;

# ls colors
LS_COLORS=$LS_COLORS:'di=1;32:ex=4;31' ; export LS_COLORS

export NVM_DIR="$HOME/.nvm"
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"  # This loads nvm
[ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"  # This loads nvm bash_completion
source /opt/homebrew/opt/powerlevel10k/powerlevel10k.zsh-theme

# Added by Antigravity
export PATH="/Users/cole/.antigravity/antigravity/bin:$PATH"

# Created by `pipx` on 2025-12-14 21:03:00
export PATH="$PATH:/Users/cole/.local/bin"
