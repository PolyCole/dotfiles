package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dots"
)

func main() {
	dotfiles := os.Getenv("DOTFILES")
	if dotfiles == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "dots: cannot determine home directory:", err)
			os.Exit(1)
		}
		dotfiles = filepath.Join(home, "repos", "dotfiles")
	}

	machine := os.Getenv("DOTFILES_MACHINE")

	// Collect module paths
	paths, err := filepath.Glob(filepath.Join(dotfiles, "modules", "*.zsh"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "dots: glob error:", err)
		os.Exit(1)
	}
	if machine != "" {
		machinePaths, _ := filepath.Glob(filepath.Join(dotfiles, "machines", machine, "modules", "*.zsh"))
		paths = append(paths, machinePaths...)
	}

	modules := dots.ParseFiles(paths)

	// Inject the built-in sync module so it appears in the overview.
	modules = append(modules, dots.SyncModule())

	args := os.Args[1:]

	// -i flag explicitly launches interactive mode
	if len(args) == 1 && args[0] == "-i" {
		if err := dots.RunInteractive(modules); err != nil {
			fmt.Fprintln(os.Stderr, "dots:", err)
			os.Exit(1)
		}
		return
	}

	switch {
	case len(args) == 0:
		dots.RenderOverview(os.Stdout, modules)

	case args[0] == "--all":
		dots.RenderAll(os.Stdout, modules)

	case args[0] == "--search":
		term := ""
		if len(args) > 1 {
			term = args[1]
		}
		dots.RenderSearch(os.Stdout, modules, term)

	case args[0] == "sync":
		sub := ""
		if len(args) >= 2 {
			sub = args[1]
		}
		switch sub {
		case "", "status":
			if err := dots.RunSyncStatus(os.Stdout, dotfiles, machine); err != nil {
				os.Exit(1)
			}
		case "link":
			if err := dots.RunSyncLink(os.Stdout, dotfiles, machine); err != nil {
				os.Exit(1)
			}
		case "now":
			if err := dots.RunSyncNow(os.Stdout, dotfiles, machine); err != nil {
				os.Exit(1)
			}
		case "install":
			if err := dots.RunSyncInstall(os.Stdout, dotfiles, machine); err != nil {
				os.Exit(1)
			}
		case "uninstall":
			if err := dots.RunSyncUninstall(os.Stdout); err != nil {
				os.Exit(1)
			}
		case "--help", "-h":
			fmt.Fprintln(os.Stdout, "Usage: dots sync [subcommand]")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Manage dotfile synchronization via launchd.")
			fmt.Fprintln(os.Stdout, "")
			fmt.Fprintln(os.Stdout, "Subcommands:")
			fmt.Fprintln(os.Stdout, "  status      Show sync daemon status and last run info (default)")
			fmt.Fprintln(os.Stdout, "  install     Install and start the launchd sync agent")
			fmt.Fprintln(os.Stdout, "  uninstall   Stop and remove the launchd sync agent")
			fmt.Fprintln(os.Stdout, "  now         Run a sync immediately")
			fmt.Fprintln(os.Stdout, "  link        Re-link dotfiles to their targets")
		default:
			fmt.Fprintf(os.Stderr, "dots sync: unknown subcommand %q. Run 'dots sync --help' to see available subcommands.\n", sub)
			os.Exit(1)
		}

	default:
		// dots <group>
		group := args[0]
		for _, m := range modules {
			if m.Name == group {
				dots.RenderGroup(os.Stdout, m)
				return
			}
		}
		fmt.Fprintf(os.Stderr, "dots: unknown group %q. Run 'dots' to see all groups.\n", group)
		os.Exit(1)
	}
}
