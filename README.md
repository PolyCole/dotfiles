# Cole's Dotfiles

![D.O.T.S.](assets/dots.png)

Commands live in self-documenting modules, discoverable via the `dots` command. The system is intentionally simple, just zsh files with structured comment blocks.

## Directory Structure

```
shell/        Core startup files (init.zsh, path.zsh, dots.zsh)
modules/      Shared zsh modules loaded on all machines
machines/     Machine-specific config (personal, ibotta, unknown)
tools/        Standalone utility programs (Go binaries, etc.)
archive/      Deprecated configs kept for reference, not sourced
bin/          Compiled binaries (startup-message)
```

### modules/

Each `.zsh` file is a self-contained group of related shell functions and aliases. Every module must have a `# Commands:` block near the top — a series of comment lines describing each command. This block is parsed by the `dots` discovery system.

```zsh
# modules/example.zsh
# Brief description of the module

# Commands:
#   my-cmd <arg>    Do something useful
#   other-cmd       Another command
```

### machines/

Machine directories are named after logical profiles, not hostnames. Machine detection runs in `shell/init.zsh` and sets `$DOTFILES_MACHINE`.

| Profile    | Hostname pattern                |
|------------|---------------------------------|
| `personal` | `six`                           |
| `ibotta`   | Matches `^[A-Z]{1}[A-Z0-9-]+$` |
| `unknown`  | Fallback                        |

Each machine directory may contain: `init.zsh`, `path.zsh`, `aliases.zsh`, `gitconfig`, `Brewfile`, and a `modules/` subdirectory for machine-specific modules.

## Setup

One command bootstraps a fresh machine (builds binaries, links configs, installs hooks — idempotent, safe to re-run):

```bash
git clone git@github.com:PolyCole/dotfiles.git ~/repos/dotfiles
cd ~/repos/dotfiles && ./install.sh
```

## The `dots` Command

`dots` surfaces available commands by parsing `# Commands:` blocks from all loaded modules. Tab completion for groups and subcommands is built in.

```bash
dots                      # Overview: all groups and command counts
dots <group>              # Detail for one group (e.g. dots git)
dots --all                # Every command across all groups
dots --search <term>      # Search command descriptions
dots edit <group>         # Open a group's module file in $EDITOR
dots doctor               # Lint modules, sync.yml, and hooks.conf for drift
```

Modules from `modules/*.zsh` are always loaded. If `$DOTFILES_MACHINE` is set, `machines/$DOTFILES_MACHINE/modules/*.zsh` is also loaded.

## Adding Commands

1. Run `dots edit <group>` (or create a new `.zsh` file in `modules/`).
2. Add the function or alias, and a matching entry in the `# Commands:` block.
3. Run `dots doctor` — it flags commands that are documented but not defined, and vice versa.

## Deprecating / Archiving Commands

To deprecate a command, remove it from the module and its `# Commands:` entry. If you want to keep it for historical reference, move the file (or relevant portion) to `archive/`. Files in `archive/` are never sourced.

## Adding a New Machine

1. Create `machines/<profile>/` with any of: `init.zsh`, `path.zsh`, `aliases.zsh`, `modules/`, `sync.yml`, `hooks.conf`.
2. Add a detection case for its hostname in `shell/init.zsh`.
3. Run `dots sync link` to symlink its configs, and `dots sync install` for automated sync.

Until a machine is recognized, it falls back to the `unknown` profile, which prints instructions instead of failing.

## Syncing Changes

The `dots sync` subsystem manages config symlinks and periodic snapshots, driven by each machine's `sync.yml`:

```bash
dots sync status          # Launchd job state, symlink health, snapshot ages
dots sync link            # Create/update symlinks declared in sync.yml
dots sync now             # Pull, rebuild binaries, snapshot, commit, and push
dots sync install         # Rebuild binaries and install the launchd agent (runs daily at 09:00)
dots sync uninstall       # Remove the launchd agent
```

When a sync pulls new commits it rebuilds `bin/dots` and `bin/startup-message`, so machines never run stale binaries. If an unattended sync fails, a macOS notification is sent instead of the failure vanishing into the log.

If you previously used `backup_dotfiles.sh` as a cron job, remove that entry from your crontab (`crontab -e`).

## Building the Binaries

`bin/` is gitignored — build the Go tools once per machine:

```bash
make            # builds bin/startup-message and bin/dots
make test       # vet + tests for the dots tool
make clean      # removes the binaries
```
