#!/bin/sh

# Create a timestamp alias for the commit message.
timestamp() {
  date +"%Y-%m-%d @ %T"
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
cp /home/cole/.bashrc /home/cole/repos/dotfiles/linux_dotfiles/bashrc
cp /home/cole/.bash_logout /home/cole/repos/dotfiles/linux_dotfiles/bash_logout
cp /home/cole/.gitconfig /home/cole/repos/dotfiles/linux_dotfiles/gitconfig
cp /home/cole/.cole.theme.bash /home/cole/repos/dotfiles/linux_dotfiles/cole_theme


cd /home/cole/repos/dotfiles/linux_dotfiles

# Finally, let's pull and push the repo.
git_changes=`git status --porcelain`
blank=""

if [ git_changes != blank ]; then
  git pull origin main
  git add .
  git commit -m "Update: $(timestamp)"
  git push origin main
else
  echo "No changes detected. You're backed-up!"
fi
