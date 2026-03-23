# Commands:
#   git-purge-dir <dir>    Remove directory from repo history (filter-branch)
#   git-purge-file <file>  Remove file from repo history (filter-branch)
#   git-amend <message>    Amend previous commit with new message

# Purges a directory from a git repository, including the repo's history.
git-purge-dir() {
  if [ $# -ne 1 ]; then
    print "Please specify the directory to purge using: \n"
    print "git-purge-dir [directory]\n"
    exit 0
  fi;

  git filter-branch --tree-filter "rm -rf $1" --prune-empty HEAD
  git for-each-ref --format="%(refname)" refs/original/ | xargs -n 1 git update-ref -d
  echo $1/ >> .gitignore
  git add .gitignore
  git commit -m "Removing $1 from git history."
  git gc
  git push origin main --force
}

# Purges a file from a git repository, including the repo's history.
git-purge-file() {
  if [ $# -ne 1 ]; then
    print "Please specify the file to purge using: \n"
    print "git-purge-file [filename]\n"
    exit 0
  fi;

  git filter-branch --tree-filter "rm -f $1" --prune-empty HEAD
  git for-each-ref --format="%(refname)" refs/original/ | xargs -n 1 git update-ref -d
  echo $1/ >> .gitignore
  git add .gitignore
  git commit -m "Removing $1 from git history."
  git gc
  git push origin main --force
}

# Amends the previous git commit message with the message passed in.
function git-amend() {
  if [ $# -ne 1 ]; then
    print "Please specify the new commit message using: \n"
    print "git-amend [new commit message]\n"
  else
    git commit --amend -m $1
  fi
}
