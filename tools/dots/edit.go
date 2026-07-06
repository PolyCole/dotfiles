package dots

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunEdit opens the module file for the given group in the user's editor.
// $EDITOR is respected (and may contain arguments, e.g. "code --wait"),
// falling back to $VISUAL and then vi.
func RunEdit(modules []Module, group string) error {
	for _, m := range modules {
		if m.Name != group {
			continue
		}
		if m.Path == "" {
			return fmt.Errorf("%q is a built-in group with no module file", group)
		}

		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vi"
		}

		parts := strings.Fields(editor)
		args := append(parts[1:], m.Path)
		cmd := exec.Command(parts[0], args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	return fmt.Errorf("unknown group %q — run 'dots' to see all groups", group)
}
