# machines/personal/init.zsh — personal machine bootstrap

# ---------------------------------------------------------------------------
# Startup banner
# ---------------------------------------------------------------------------
BACKGROUND=$(( $RANDOM % 219 + 22 ))
COLOR="\e[38;5;231m\e[48;5;${BACKGROUND}m"
NC='\033[0m'

_indent="\t\t"
echo ""
echo "${_indent}${COLOR}                            ${NC}"
echo "${_indent}${COLOR}             _              ${NC}"
echo "${_indent}${COLOR}            (_)             ${NC}"
echo "${_indent}${COLOR}       ___   _  __  __      ${NC}"
echo "${_indent}${COLOR}      / __| | | \\ \\/ /      ${NC}"
echo "${_indent}${COLOR}      \\__ \\ | |  >  <       ${NC}"
echo "${_indent}${COLOR}      |___/ |_| /_/\\_\\      ${NC}"
echo "${_indent}${COLOR}                            ${NC}"
echo ""

COLUMNS=28 PADDING=16 $DOTFILES/bin/startup-message $DOTFILES/machines/personal/messages.txt
unset _indent

