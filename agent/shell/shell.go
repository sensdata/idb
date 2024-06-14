package shell

import (
	"bytes"
	"os/exec"
)

func ExecuteCommand(command string) (string, error) {
	// TODO: check prefix in command to excute different commands
	// For now, using 'sh' command for test
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
