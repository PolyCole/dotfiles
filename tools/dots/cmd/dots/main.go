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

	case args[0] == "sync" && len(args) >= 2 && args[1] == "link":
		if err := dots.RunSyncLink(os.Stdout, dotfiles, machine); err != nil {
			os.Exit(1)
		}

	case args[0] == "sync" && len(args) >= 2 && args[1] == "status":
		if err := dots.RunSyncStatus(os.Stdout, dotfiles, machine); err != nil {
			os.Exit(1)
		}

	case args[0] == "sync" && len(args) >= 2 && args[1] == "now":
		if err := dots.RunSyncNow(os.Stdout, dotfiles, machine); err != nil {
			os.Exit(1)
		}

	case args[0] == "sync" && len(args) >= 2 && args[1] == "install":
		if err := dots.RunSyncInstall(os.Stdout, dotfiles, machine); err != nil {
			os.Exit(1)
		}

	case args[0] == "sync" && len(args) >= 2 && args[1] == "uninstall":
		if err := dots.RunSyncUninstall(os.Stdout); err != nil {
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
