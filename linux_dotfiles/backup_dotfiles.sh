#!/bin/bash

# Create a timestamp alias for the commit message.
timestamp() {
  date +"%Y-%m-%d @ %T"
}

# Selects a random emoji to be used with the commit message.
randomEmoji() {
  emojis=("🟠" "🟣" "⚫" "🟡" "🔵" "🟤" "🟢" "⚪" "⭕" "🔴" "⏺️")
  selected_emoji=${emojis["$[RANDOM % ${#emojis[@]}]"]}
}

ssh -T -i ~/.ssh/id_rsa git@github.com

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
cp /home/cole/.bashrc /home/cole/repos/dotfiles/linux_dotfiles/bashrc
cp /home/cole/.bash_logout /home/cole/repos/dotfiles/linux_dotfiles/bash_logout
cp /home/cole/.gitconfig /home/cole/repos/dotfiles/linux_dotfiles/gitconfig
cp /home/cole/.cole.theme.bash /home/cole/repos/dotfiles/linux_dotfiles/cole_theme
cp /home/cole/.cole_config /home/cole/repos/dotfiles/linux_dotfiles/cole_config


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