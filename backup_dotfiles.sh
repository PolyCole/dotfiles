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

if [ "$HOST" = "six" ]; then
  # Grabbing Brewfile.
  cd ~
  brew bundle dump
  cp Brewfile $DOTFILES/personal
  rm Brewfile

  # Grabbing global python packages.
  cd ~
  python -m pip list >> python_packages
  cp python_packages $DOTFILES/personal
  rm python_packages

  cp $HOME/.gitconfig $DOTFILES/personal/gitconfig
else
  # Grabbing Brew Packages
  cd ~
  /usr/local/bin/brew bundle dump
  cp Brewfile $DOTFILES/work
  rm Brewfile

  # Grabbing some other dotfiles.
  cp $HOME/.asdfrc $DOTFILES/work/asdfrc
  cp $HOME/.gitconfig $DOTFILES/work/gitconfig
  cp $HOME/.p10k.zsh $DOTFILES/work/p10k.zsh
  cp $HOME/.linker_aliases $DOTFILES/work/linker_aliases

  # Grabbing globally installed npm packages
  npm -g list >> npm_list
  cp npm_list $DOTFILES/work/npm_list
  rm npm_list
fi 

# Copying top-level zshrc into place.
cp ~/.zshrc $DOTFILES/zshrc

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
