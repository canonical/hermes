package collector

import (
	"fmt"
	"os"
	"os/exec"
)

func PrepareCmd(cmds []string, env map[string]string) *exec.Cmd {
	if len(env) > 0 {
		for i, cmd := range cmds {
			if replace, isExist := env[cmd]; isExist {
				cmds[i] = replace
			}
		}
	}

	cmd := exec.Command(cmds[0], cmds[1:]...)
	cmd.Env = os.Environ()
	for key, val := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}

	return cmd
}
