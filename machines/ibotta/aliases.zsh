# machines/ibotta/aliases.zsh — General Ibotta aliases and helpers

# ---------------------------------------------------------------------------
# Clipboard / JSON formatting
# ---------------------------------------------------------------------------
alias format-clip="jq '.Message' | jq -r '.' | jq ."

# ---------------------------------------------------------------------------
# Git shortcuts
# ---------------------------------------------------------------------------
function qc() {
  if [ $# -eq 0 ]; then
    echo "Please provide a commit message:"
    echo "qc \"your commit message\""
  else
    git add -A && git commit -m "$1" && git push
  fi
}

# ---------------------------------------------------------------------------
# Schedule reminder
# ---------------------------------------------------------------------------
schedule() {
  print "\n***************************************"
  print "🟢 Dotfiles -- M, W, F   @9:30 am"
  print "🧰 Toolbox  -- T, R,     @9:30 am"
  print "🧰 Toolbox  -- F         @3:30 pm"
  print "***************************************"
}

# ---------------------------------------------------------------------------
# Repo / tool navigation
# ---------------------------------------------------------------------------
alias repos="cd /Users/cole.polyak/repos"
alias tools="~/toolbox"

# ---------------------------------------------------------------------------
# Receipt checking
# ---------------------------------------------------------------------------
alias check_receipt="python ~/toolbox/ibotta_scripts/check_receipt.py"

# ---------------------------------------------------------------------------
# Keychain helpers (zsh-osx-keychain wrappers)
# ---------------------------------------------------------------------------
alias peek-zshenv-var="keychain-environment-variable"
alias rm-zshenv-var="delete-keychain-environment-variable"
alias zshenv-var="set-keychain-environment-variable"

# ---------------------------------------------------------------------------
# SSH aliases
# ---------------------------------------------------------------------------
alias ssh-receipts="ssh receipts.dev.ibotta.com"
alias ssh-tools="ssh tools.ibotta.com"
alias ssh-cole="ssh colep.dev.ibotta.com"

# ---------------------------------------------------------------------------
# NPX / build cleanup
# ---------------------------------------------------------------------------
alias clean="npx nx run-many --target=clean --all"

# ---------------------------------------------------------------------------
# Granted/Assume ABAC
# ---------------------------------------------------------------------------
alias assume=". assume"

# ---------------------------------------------------------------------------
# Linker aliases (programmatically generated)
# ---------------------------------------------------------------------------
if [ -f $HOME/.linker_aliases ]; then
  . $HOME/.linker_aliases
fi
