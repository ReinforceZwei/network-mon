package pingwrap

import "os/exec"

type PingLinux struct{}

func (p PingLinux) PingOnce(address string) bool {
	cmd := exec.Command("ping", address, "-c", "1", "-W", "10", "-n", "-q")
	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode() == 0
		}
	}
	// Zero return code
	return true
}
