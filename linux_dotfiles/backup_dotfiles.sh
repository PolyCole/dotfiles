#!/bin/bash

# Author: Cole Polyak
# 1 August 2020
# This script facilitates the backing up of dotfiles on Five-Linux

# Create a timestamp alias for the commit message.
timestamp() {
  date +"%Y-%m-%d @ %T"
}

# Selects a random emoji to be used with the commit message.
randomEmoji() {
  emojis=("π " "π£" "β«" "π‘" "π΅" "π€" "π’" "βͺ" "β­" "π΄" "βΊοΈ")
  selected_emoji=${emojis["$[RANDOM % ${#emojis[@]}]"]}
}

# First, let's dump apt just to be sure we have that on record.
cd ~
apt list > apt-list
cp apt-list /home/cole/repos/dotfiles/linux_dotfiles/
rm apt-list

# Second, let's dump snap too.
snap list > snap-list
cp snap-list /home/cole/repos/dotfiles/linux_dotfiles/
rm snap-list

# Third, let's dump pip.
/usr/bin/python3 -m pip list > pip-list
cp pip-list /home/cole/repos/dotfiles/linux_dotfiles/
rm pip-list

# Fourth, let's dump npm.
npm ls -g --depth=0 > npm-list
cp npm-list /home/cole/repos/dotfiles/linux_dotfiles/
rm npm-list


# Fifth, let's snag some dotfiles.
cp /home/cole/.zshrc /home/cole/repos/dotfiles/linux_dotfiles/zshrc
cp /home/cole/.gitconfig /home/cole/repos/dotfiles/linux_dotfiles/gitconfig
cp /home/cole/powerlevel10k/powerlevel10k.zsh-theme /home/cole/repos/dotfiles/linux_dotfiles/cole_theme

cd /home/cole/repos/dotfiles/linux_dotfiles

# Finally, let's pull and push the repo.

# Check git status
gs="$(git status | grep -i "modified")"

# If there is a new change
if [[ $gs == *"modified"* ]]; then
  randomEmoji

  git pull origin main
  git add .
  git commit -m "$selected_emoji Update: $(timestamp)"
  git push origin main
else
  echo "No changes detected. You're backed-up!"
fi
