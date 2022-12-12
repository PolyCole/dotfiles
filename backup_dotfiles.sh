#!/bin/sh

# Create a timestamp alias for the commit message.
timestamp() {
  date +"%Y-%m-%d @ %T"
}

# Selects a random emoji to be used with the commit message.
randomEmoji() {
  emojis=("🟠" "🟣" "⚫" "🟡" "🔵" "🟤" "🟢" "⚪" "⭕" "🔴" "⏺️")
  selected_emoji=${emojis["$[RANDOM % ${#emojis[@]}]"]}
}

# We're going to ignore this for now. I need to circle back and actually make this work across machines...
# # First, let's dump homebrew.
# cd ~
# /usr/local/bin/brew bundle dump
# cp Brewfile /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles
# rm Brewfile


# # Second, let's snag some dotfiles.
# cp /Users/cole.polyak/.zshrc /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/zshrc # <----- This is one that we defnitely want...
# cp /Users/cole.polyak/.asdfrc /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/asdfrc
# cp /Users/cole.polyak/.gitconfig /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/gitconfig
# cp /Users/cole.polyak/.p10k.zsh /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/p10k.zsh
# cp /Users/cole.polyak/.linker_aliases /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/linker_aliases

# # Third, let's make sure we grab globally installed npm packages...
# npm -g list >> npm_list
# cp npm_list /Users/cole.polyak/Desktop/hub/repos/dotfiles/mac_dotfiles/npm_list
# rm npm_list

# Finally, let's pull and push the repo.
cd $HOME/repos/dotfiles

if [[ `git status --porcelain` ]]; then
  randomEmoji

  git pull origin main
  git add .
  git commit -m "$selected_emoji Update: $(timestamp)"
  git push origin main
else
  echo "********************************************"
  echo "No changes detected. Dotfiles are backed-up!"
  echo "********************************************"
fi
