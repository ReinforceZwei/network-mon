package pingwrap

import (
	"os/exec"
)

type PingWindows struct{}

func (p PingWindows) PingOnce(address string) bool {
	cmd := exec.Command("ping", address, "-n", "1", "-w", "10000")
	if err := cmd.Run(); err != nil {
		// Non-zero return code
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 0
		}
	}
	// Zero return code
	return true
}
