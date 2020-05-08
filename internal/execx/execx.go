package execx

import (
	"os/exec"
	"strings"
)

// Describe the command being executed.
func Describe(cmd *exec.Cmd) string {
	return strings.Join(cmd.Args, " ") + "\n" + strings.Join(cmd.Env, "\n")
}
