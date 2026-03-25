# machines/personal/init.zsh — personal machine bootstrap

# ---------------------------------------------------------------------------
# Startup banner
# ---------------------------------------------------------------------------
BACKGROUND=$(( $RANDOM % 219 + 22 ))
COLOR="\e[38;5;231m\e[48;5;${BACKGROUND}m"
NC='\033[0m'
echo "
        ${COLOR}                            ${NC}
\t${COLOR}             _              ${NC}
\t${COLOR}            (_)             ${NC}
\t${COLOR}       ___   _  __  __      ${NC}
\t${COLOR}      / __| | | \ \/ /      ${NC}
\t${COLOR}      \__ \ | |  >  <       ${NC}
\t${COLOR}      |___/ |_| /_/\_\      ${NC}
        ${COLOR}                            ${NC}
"

$DOTFILES/bin/startup-message $DOTFILES/machines/personal/messages.txt

