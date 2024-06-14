package shell

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func ExecuteCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", errors.New("empty command")
	}

	switch parts[0] {
	case "sh":
		cmdStr := strings.Join(parts[1:], " ")
		return executeCommand(parts[0], []string{"-c", cmdStr})
	default:
		return executeCommand(parts[0], parts[1:])
	}
}

func executeCommand(name string, args []string) (string, error) {
	cmd := exec.Command(name, args...)
	return runCommand(cmd)
}

func runCommand(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return out.String(), nil
}
