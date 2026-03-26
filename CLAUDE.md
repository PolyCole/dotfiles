# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

Personal dotfiles repository for shell configuration, synchronized across machines. Uses cron to automatically commit and push changes. Requires SSH-based GitHub access for cron compatibility.

## Directory Structure

```
shell/        Core shell startup files (init.zsh, path.zsh, dots.zsh)
modules/      Shared zsh modules loaded on all machines
machines/     Machine-specific configuration directories
tools/        Standalone utility programs (e.g. startup-message in Go)
archive/      Deprecated configs kept for reference, not sourced
```

### modules/

Each `.zsh` file in `modules/` is a self-contained set of related shell functions and aliases. Every module must have a `# Commands:` block near the top — a series of comment lines describing each command. This block is parsed by the `dots` discovery system.

Example:
```zsh
# modules/example.zsh
# Brief description of the module

# Commands:
#   my-cmd <arg>    Do something useful
#   other-cmd       Another command
```

### machines/

Machine directories are named after logical machine profiles, not hostnames. Machine detection runs in `shell/init.zsh` and sets `$DOTFILES_MACHINE`. Current profiles:

- `personal` — personal laptop (hostname: `six`)
- `ibotta` — work machines (hostnames matching `^[A-Z]{1}[A-Z0-9-]+$`)
- `unknown` — fallback for unrecognized machines

Each machine directory may contain: `init.zsh`, `path.zsh`, `aliases.zsh`, `gitconfig`, `Brewfile`, and a `modules/` subdirectory for machine-specific modules.

### archive/

Configs that are no longer active. Do not source these. Keep them for historical reference only.

## The `dots` Discovery System

`dots` is a shell function (defined in `shell/dots.zsh`) that surfaces available commands by parsing `# Commands:` blocks from all modules.

```bash
dots                      # Overview: all groups and command counts
dots <group>              # Detail for one group (e.g. dots git)
dots --all                # Every command across all groups
dots --search <term>      # Search command descriptions
```

It loads modules from `modules/*.zsh` and, if `$DOTFILES_MACHINE` is set, also from `machines/$DOTFILES_MACHINE/modules/*.zsh`.

## Environment Variables

- `$DOTFILES` — Points to this repository
- `$DOTFILES_MACHINE` — Active machine profile (`personal`, `ibotta`, `unknown`)
- `$HOME/.linker_aliases` — Dynamically generated aliases (sourced if exists)

# Agent Instructions

This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.

## Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
bd sync               # Sync with git
```

### Tool Usage Tips
- When running `bd doctor`, warnings are far less important than errors. Errors should be addressed
  immediately.
- `bd doctor` warnings should be communicated to the user, but largely ignored.
- Prefer using standalone issues as opposed to dot-notation sub-issues.

## Landing the Plane (Session Completion)

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   bd sync
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
- Do not include "co-authored by" in your commit messages.
- Do not write commit messages longer than a sentence.
- Do not include more detail than necessary in commit messages, they should be short and to the point.
