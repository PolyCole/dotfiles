# Archived git functions
# Retired code — not sourced anywhere.

# Archived: 2026-03-22
# Reason: push.autoSetupRemote makes this obsolete
function git-upstream() {
  BRANCH_NAME=$(git branch --show-current)
  git push --set-upstream origin $BRANCH_NAME
}
