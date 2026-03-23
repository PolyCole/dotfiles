# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

Personal dotfiles repository for shell configuration backup and synchronization across machines. Uses cron to automatically commit and push changes. Requires SSH-based GitHub access for cron compatibility.

## Structure

- `common/` - Shared configuration files sourced across all machines
  - `.toolbox_config` - Toolbox aliases and programmatic linker alias loading
  - `llm_instructions.md` - LLM prompting guidelines (Planner Mode, Debugger Mode)
  - `parker_claude_code_mode.md` - AI interaction directives (Forklift vs Weightlifting tasks, risk analysis, objective truth)

## Key Shell Functions

- `git-purge-dir/git-purge-file` - Removes files/directories from git history (uses filter-branch)
- `murderdocker` - Aggressive Docker cleanup (volumes, containers, images, networks)
- `mkd` - Create directory and cd into it
- `weather [location]` - Fetch weather (defaults to Denver)
- `aoc [day]` - Advent of Code template generator

## Environment Variables

- `$DOTFILES` - Points to this repository
- `$HOME/.linker_aliases` - Dynamically generated aliases (sourced if exists)

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
- Prefer using standlone issues as opposed to dot-notation sub-issues.

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
