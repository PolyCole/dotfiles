# modules/navigation.zsh
# Filesystem navigation and directory management utilities

# Commands:
#   o [path]       Open current or specified directory in Finder
#   a [path]       Open current or specified directory in VS Code
#   mkd <dir>      Create directory and cd into it
#   repos          Alias to ~/repos
#   dotfiles       Alias to $DOTFILES

alias repos="$HOME/repos"
alias dotfiles="$DOTFILES"

alias a="a "
alias o="o "

# Opens the current (or specified) directory in file explorer.
function o() {
	if [ $# -eq 0 ]; then
		open .;
	else
		open "$@";
	fi;
}

# Opens the current (or specified) directory in atom text editor.
function a() {
	if [ $# -eq 0 ]; then
		code .;
	else
		code "$@";
	fi;
}

# Create a new directory and enter it
mkd() {
	mkdir -p "$@" && cd "$_";
}
