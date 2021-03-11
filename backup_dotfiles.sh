#!/bin/sh

# Create a timestamp alias for the commit message.
timestamp() {
  date +"%Y-%m-%d @ %T"
}

# First, let's dump homebrew.
cd ~
brew bundle dump
cp Brewfile /Users/cole.polyak/Desktop/hub/repos/dotfiles/
rm Brewfile


# Second, let's snag some dotfiles.
cp /Users/cole.polyak/.zshrc /Users/cole.polyak/Desktop/hub/repos/dotfiles/zshrc
cp /Users/cole.polyak/.asdfrc /Users/cole.polyak/Desktop/hub/repos/dotfiles/asdfrc
cp /Users/cole.polyak/.gitconfig /Users/cole.polyak/Desktop/hub/repos/dotfiles/gitconfig

cd /Users/cole.polyak/Desktop/hub/repos/dotfiles

# Finally, let's pull and push the repo.
if [[ `git status --porcelain` ]]; then
  git pull origin main
  git add .
  git commit -m "pdate: $(timestamp)"
  git push origin main
fi
