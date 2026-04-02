# Work Machine Onboarding Guide

This document walks through setting up the D.O.T.S. (Don't Overthink This Shit) system on your work machine. It assumes you're coming from the old flat config structure (`work/.work_config`, `common/.common_config`, etc.) and moving to the new modular layout.

---

## What Changed

The old structure sourced monolithic config files per machine. The new structure is:

```
zshrc                          Entry point (symlinked to ~/.zshrc)
shell/init.zsh                 Machine detection, oh-my-zsh, p10k
shell/path.zsh                 Homebrew PATH setup
shell/dots.zsh                 dots command wrapper
modules/*.zsh                  Shared modules (all machines)
machines/ibotta/init.zsh       Ibotta banner, env vars, keychain
machines/ibotta/path.zsh       rbenv, NVM, toolbox PATHs
machines/ibotta/aliases.zsh    Work-specific aliases
machines/ibotta/modules/*.zsh  Work-specific modules (aws, deploy, k8s, etc.)
machines/ibotta/sync.yml       Symlink + snapshot manifest
machines/ibotta/hooks.conf     Git hook declarations
```

Everything that was in `.work_config` and `.common_config` has been extracted into these files. Nothing was removed -- just reorganized and made discoverable via `dots`.

---

## Prerequisites

- **Homebrew** installed
- **Go** installed (for building binaries): `brew install go`
- **oh-my-zsh** installed at `~/.oh-my-zsh`
- **powerlevel10k** available as oh-my-zsh plugin or via Homebrew
- Repo cloned to `~/repos/dotfiles`

---

## Setup Steps

### 1. Pull the branch

```bash
cd ~/repos/dotfiles
git fetch origin
git checkout factory-aurora   # or main, once merged
```

### 2. Build the binaries

Two Go programs need to be compiled:

```bash
make all
```

This builds:
- `bin/startup-message` -- prints a random message from `machines/ibotta/messages.txt` at shell startup
- `bin/dots` -- the discovery and sync tool

Verify both exist:
```bash
ls -la bin/startup-message bin/dots
```

### 3. Symlink your zshrc

The most critical symlink -- everything else flows from here:

```bash
ln -sf ~/repos/dotfiles/zshrc ~/.zshrc
```

Or use `dots sync link` once you have a working shell (chicken-and-egg -- do this manually first).

### 4. Symlink remaining configs

These are declared in `machines/ibotta/sync.yml` and can be linked manually or via `dots sync link`:

| Source | Target |
|--------|--------|
| `machines/ibotta/gitconfig` | `~/.gitconfig` |
| `machines/ibotta/asdfrc` | `~/.asdfrc` |
| `machines/ibotta/p10k.zsh` | `~/.p10k.zsh` |
| `machines/ibotta/linker_aliases` | `~/.linker_aliases` |
| `zshrc` | `~/.zshrc` |

To link all at once (after your shell is working):

```bash
dots sync link
```

### 5. Open a new shell

```bash
exec zsh
```

You should see:
1. The IBOTTA ASCII banner
2. A random startup message
3. Your normal prompt

If the `dots` command errors about a missing binary, re-run `make all`.

### 6. Verify machine detection

```bash
echo $DOTFILES_MACHINE
# Should print: ibotta
```

Machine detection in `shell/init.zsh` matches uppercase hostnames (e.g., `COLEP-MBP-1`) to the `ibotta` profile. If your hostname doesn't match `^[A-Z]{1}[A-Z0-9-]+$`, you'll get `unknown` instead.

### 7. Install git hooks

```bash
dots-hooks-install
```

This creates dispatcher scripts in `~/.git-hooks/` and sets `core.hooksPath` globally. The hooks are declared in `machines/ibotta/hooks.conf` and include:

- `pre-commit` -- beads issue tracking + trufflehog secret scanning (if installed)
- `post-checkout`, `post-merge`, `pre-push`, `prepare-commit-msg` -- beads integration

### 8. Set up keychain credentials (if not already present)

`machines/ibotta/init.zsh` loads these from macOS Keychain at shell startup:

```bash
# Check what's already stored:
security find-generic-password -s "GEM_REPO_LOGIN" 2>/dev/null && echo "found" || echo "missing"
security find-generic-password -s "NPM_REPO_LOGIN" 2>/dev/null && echo "found" || echo "missing"
security find-generic-password -s "MVN_REPO_LOGIN" 2>/dev/null && echo "found" || echo "missing"
security find-generic-password -s "GEMINI_API_KEY" 2>/dev/null && echo "found" || echo "missing"
```

If missing, add them:
```bash
security add-generic-password -a "$USER" -s "GEM_REPO_LOGIN" -w "your-token"
# repeat for others as needed
```

### 9. (Optional) Install automated sync

Sets up a launchd agent to periodically sync snapshots (Brewfile, npm list):

```bash
dots sync install
```

Check status anytime with:
```bash
dots sync status
```

---

## Load Order Reference

When you open a terminal, files source in this order:

1. `zshrc` -- sets `$DOTFILES`, sources everything below
2. `shell/init.zsh` -- detects machine, loads oh-my-zsh + p10k
3. `shell/path.zsh` -- Homebrew + GNU coreutils PATH
4. `shell/dots.zsh` -- defines `dots()` shell function
5. `modules/*.zsh` -- all shared modules (alphabetical)
6. `machines/ibotta/path.zsh` -- rbenv, NVM, toolbox PATHs
7. `machines/ibotta/init.zsh` -- banner, env vars, keychain loading
8. `machines/ibotta/aliases.zsh` -- work aliases

---

## Using dots

Once set up, `dots` is your command discovery tool:

```bash
dots                    # Overview of all groups and command counts
dots git                # Show git module commands
dots deploy             # Show deploy module commands
dots --all              # Every command across all groups
dots --search proxy     # Search for commands matching "proxy"
```

It parses the `# Commands:` blocks in every module file, so any new command you add to a module is automatically discoverable.

---

## Sync System

The sync system manages config symlinks and periodic snapshots.

```bash
dots sync status        # Current state of all managed configs and snapshots
dots sync link          # Create/update all symlinks from sync.yml
dots sync now           # Run all snapshot commands immediately
dots sync install       # Install launchd agent for periodic sync
dots sync uninstall     # Remove launchd agent
```

Snapshots defined in `machines/ibotta/sync.yml`:
- `brew bundle dump` --> `machines/ibotta/Brewfile`
- `npm -g list` --> `machines/ibotta/npm_list`

---

## Modules on Your Work Machine

### Shared (all machines)
| Module | What it provides |
|--------|-----------------|
| `aliases` | General aliases (ll, la, src, etc.) |
| `git` | Git utilities (gca, gp, nuke, stash helpers) |
| `docker` | Docker shortcuts (dps, dex, etc.) |
| `navigation` | Directory nav (mkcd, up, etc.) |
| `web` | URL/HTTP utilities |
| `crypto` | Crypto price checks |
| `advent` | Advent of Code helpers |
| `hooks` | Git hooks management (dots-hooks-install) |

### Ibotta-specific
| Module | What it provides |
|--------|-----------------|
| `aws` | AWS/SAML2 auth, MFA helpers |
| `deploy` | Deployment workflows, PR management |
| `kubernetes` | Cluster context switching |
| `oncall` | On-call runbook aliases |
| `services` | Microservice navigation (cd-to shortcuts) |

Run `dots --all` to see every available command.

---

## Troubleshooting

**`dots: command not found`** -- The shell function in `shell/dots.zsh` couldn't find `bin/dots`. Run `make all`.

**Machine detected as `unknown`** -- Your hostname doesn't match the ibotta pattern. Check `hostname` and update the regex in `shell/init.zsh` if needed.

**Keychain errors at startup** -- Missing credentials in macOS Keychain. See step 8 above. These are non-fatal; the shell still works.

**Old configs still being sourced** -- Check for leftover `~/.zshrc` content that sources old paths. The new `zshrc` should be the only thing at `~/.zshrc`.

**`dots sync link` shows conflicts** -- A target file already exists and isn't a symlink to the expected source. Back it up and re-run, or manually replace it.
