#!/bin/sh
# install.sh — one-shot, idempotent bootstrap for a fresh machine.
#
#   ./install.sh
#
# Builds the Go binaries, symlinks configs from the machine's sync.yml,
# and installs the git hooks. Safe to re-run at any time.
# Automated daily sync is opt-in afterwards: 'dots sync install'.

set -eu

DOTFILES="$(cd "$(dirname "$0")" && pwd)"
export DOTFILES

say()  { printf '\033[1;32m==>\033[0m %s\n' "$1"; }
warn() { printf '\033[1;33mwarn:\033[0m %s\n' "$1" >&2; }
die()  { printf '\033[1;31merror:\033[0m %s\n' "$1" >&2; exit 1; }

# ---------------------------------------------------------------------------
# Prerequisites
# ---------------------------------------------------------------------------
say "Checking prerequisites"
for tool in git zsh make go; do
    command -v "$tool" >/dev/null 2>&1 || die "$tool is required — install it and re-run"
done
command -v brew >/dev/null 2>&1 || warn "Homebrew not found — shell/path.zsh expects it on macOS"
[ -d "$HOME/.oh-my-zsh" ] || warn "oh-my-zsh not found at ~/.oh-my-zsh — shell/init.zsh expects it"

# ---------------------------------------------------------------------------
# Machine detection (mirrors shell/init.zsh)
# ---------------------------------------------------------------------------
host="$(hostname -s)"
if [ -n "${DOTFILES_MACHINE:-}" ]; then
    : # already set in the environment — respect the override
elif [ "$host" = "six" ]; then
    DOTFILES_MACHINE="personal"
elif echo "$host" | grep -Eq '^[A-Z][A-Z0-9-]+$'; then
    DOTFILES_MACHINE="ibotta"
else
    DOTFILES_MACHINE="unknown"
    warn "unrecognized hostname '$host' — using the 'unknown' profile."
    warn "add a detection case to shell/init.zsh (and install.sh) for this machine."
fi
export DOTFILES_MACHINE
say "Machine profile: $DOTFILES_MACHINE"

# ---------------------------------------------------------------------------
# Build binaries
# ---------------------------------------------------------------------------
say "Building bin/dots and bin/startup-message"
make -C "$DOTFILES" all

# ---------------------------------------------------------------------------
# Symlink zshrc first (everything else flows from it), then the rest
# ---------------------------------------------------------------------------
say "Linking ~/.zshrc"
if [ -e "$HOME/.zshrc" ] && [ ! -L "$HOME/.zshrc" ]; then
    mv "$HOME/.zshrc" "$HOME/.zshrc.bak"
    warn "existing ~/.zshrc backed up to ~/.zshrc.bak"
fi
ln -sfn "$DOTFILES/zshrc" "$HOME/.zshrc"

if [ -f "$DOTFILES/machines/$DOTFILES_MACHINE/sync.yml" ]; then
    say "Linking configs from machines/$DOTFILES_MACHINE/sync.yml"
    "$DOTFILES/bin/dots" sync link
else
    warn "no sync.yml for '$DOTFILES_MACHINE' — skipping config links"
fi

# ---------------------------------------------------------------------------
# Git hooks (reuses dots-hooks-install so there is one source of truth)
# ---------------------------------------------------------------------------
say "Installing git hooks"
zsh -c "source '$DOTFILES/modules/hooks.zsh' && dots-hooks-install"

# ---------------------------------------------------------------------------
# Done
# ---------------------------------------------------------------------------
say "Done. Next steps:"
printf '    exec zsh            # start using the new shell\n'
printf '    dots                # discover available commands\n'
printf '    dots sync install   # optional: daily automated sync (macOS launchd)\n'
