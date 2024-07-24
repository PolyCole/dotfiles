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
current_hostname=$(hostname -s)

# Define the string you want to compare with
compare_string="six"

# Compare the two strings
if [ "$current_hostname" = "$compare_string" ]; then
    source $DOTFILES/personal/.personal_config
else
    source $DOTFILES/work/.work_config
fi

# ls colors
LS_COLORS=$LS_COLORS:'di=1;32:ex=4;31' ; export LS_COLORS
