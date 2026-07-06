package dots

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// notifySyncFailure sends a best-effort macOS desktop notification.
// It only fires on darwin when stdout is not a terminal — i.e. unattended
// launchd runs, where a failure would otherwise vanish into the log file.
func notifySyncFailure(msg string) {
	if runtime.GOOS != "darwin" {
		return
	}
	if fi, err := os.Stdout.Stat(); err == nil && fi.Mode()&os.ModeCharDevice != 0 {
		return // interactive run — the user already sees the error
	}
	script := fmt.Sprintf("display notification %q with title %q", msg, "dots sync failed")
	_ = exec.Command("osascript", "-e", script).Run()
}
