# modules/aliases.zsh
# General-purpose aliases and miscellaneous shell utilities

# Commands:
#   cat        → bat (syntax-highlighted cat)
#   lsa        List all files including hidden
#   lsal       List all files in long format
#   python     → python3
#   reload     Re-source ~/.zshrc
#   1:1        Print hello
#   week       Print current week of the year
#   message    Append a startup message to machines/personal/messages.txt

alias cat="bat"
alias lsa="ls -a"
alias lsal="ls -al"
alias python="python3"
alias reload="source ~/.zshrc"
alias 1:1="echo hello"

# Prints the current week of the year
week() {
  echo "Week `date +%V` of 52"
}

# For my personal config, I maintain a list of startup messages.
# I'd love to add to them regardless of what machine I'm on :)
function message() {
    echo "$1" >> $DOTFILES/machines/personal/messages.txt
}
