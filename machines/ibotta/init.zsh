# machines/ibotta/init.zsh — Ibotta machine bootstrap

# ---------------------------------------------------------------------------
# Startup banner
# ---------------------------------------------------------------------------
COLOR='\e[38;5;231m\e[48;5;198m'
NC='\033[0m'

_indent="\t "
echo ""
echo "${_indent}${COLOR}                                        ${NC}"
echo "${_indent}${COLOR}  /\$\$\$\$\$\$ /\$\$\$\$\$\$\$  /\$\$\$\$\$\$\$\$ /\$\$\$\$\$\$   ${NC}"
echo "${_indent}${COLOR} |_  \$\$_/| \$\$__  \$\$|__  \$\$__//\$\$__  \$\$  ${NC}"
echo "${_indent}${COLOR}   | \$\$  | \$\$  \ \$\$   | \$\$  | \$\$  \ \$\$  ${NC}"
echo "${_indent}${COLOR}   | \$\$  | \$\$\$\$\$\$\$    | \$\$  | \$\$\$\$\$\$\$\$  ${NC}"
echo "${_indent}${COLOR}   | \$\$  | \$\$__  \$\$   | \$\$  | \$\$__  \$\$  ${NC}"
echo "${_indent}${COLOR}   | \$\$  | \$\$  \ \$\$   | \$\$  | \$\$  | \$\$  ${NC}"
echo "${_indent}${COLOR}  /\$\$\$\$\$\$| \$\$\$\$\$\$\$/   | \$\$  | \$\$  | \$\$  ${NC}"
echo "${_indent}${COLOR} |______/|_______/    |__/  |__/  |__/  ${NC}"
echo "${_indent}${COLOR}                                        ${NC}"
echo ""

if [[ -x "$DOTFILES/bin/startup-message" ]]; then
  COLUMNS=40 PADDING=9 $DOTFILES/bin/startup-message $DOTFILES/machines/ibotta/messages.txt
else
  echo "startup-message binary not found — run 'make all' in \$DOTFILES" >&2
fi
echo ""
unset _indent

# ---------------------------------------------------------------------------
# Common locations
# ---------------------------------------------------------------------------
export REPO_BASE_PATH=~/repos

# ---------------------------------------------------------------------------
# Environment variables
# ---------------------------------------------------------------------------
export AWS_DEFAULT_REGION=us-east-1
export AWS_REGION=$AWS_DEFAULT_REGION
export AWS_SESSION_TTL=12h
export AWS_FEDERATION_TOKEN_TTL=12h
export AWS_ASSUME_ROLE_TTL=1h

export GRANTED_ALIAS_CONFIGURED="true"

export DOCKER_HOST="unix:///${HOME}/.colima/docker.sock"

export DATABRICKS_CONFIG_PROFILE=ibotta-engineering

# ---------------------------------------------------------------------------
# Keychain credential loading
# ---------------------------------------------------------------------------
export GEM_REPO_LOGIN=$(keychain-environment-variable PRIVATE_PACKAGE_REPO_LOGIN)
export NPM_REPO_LOGIN=$(keychain-environment-variable PRIVATE_PACKAGE_REPO_LOGIN)
export PY_REPO_LOGIN=$(keychain-environment-variable PY_REPO_LOGIN)
export MVN_REPO_LOGIN=$(keychain-environment-variable PRIVATE_PACKAGE_REPO_LOGIN)
export GEMINI_API_KEY=$(keychain-environment-variable GEMINI_API_KEY)

# ---------------------------------------------------------------------------
# Miscellaneous
# ---------------------------------------------------------------------------
typeset -g POWERLEVEL9K_INSTANT_PROMPT=quiet
