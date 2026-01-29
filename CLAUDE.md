# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository Purpose

Personal dotfiles repository for shell configuration backup and synchronization across machines. Uses cron to automatically commit and push changes. Requires SSH-based GitHub access for cron compatibility.

## Structure

- `common/` - Shared configuration files sourced across all machines
  - `.common_config` - Aliases and shell functions (cat→bat, docker helpers, git utilities, weather, etc.)
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
