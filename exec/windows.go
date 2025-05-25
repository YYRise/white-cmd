//go:build windows
// +build windows

package exec

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ExecSh(line, dir string) (string, error) {
	var shell string
	if shell = os.Getenv("COMSPEC"); shell == "" {
		shell = "cmd"
	}
	cmd := exec.Command(shell, "/c", line)
	if dir != "" {
		cmd.Dir = dir
	}
	b, err := cmd.Output()
	if err != nil {
		if eerr, ok := err.(*exec.ExitError); ok {
			b = eerr.Stderr
		}
		return "", fmt.Errorf("%s: %w", string(b), err)
	}
	return strings.TrimSpace(string(b)), nil
}
