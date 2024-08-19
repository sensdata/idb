package shell

import (
	"bytes"
	"fmt"
	"os/exec"
)

func ExecuteCommandIgnore(command string) error {
	output, err := ExecuteCommand(command)
	if err != nil {
		return fmt.Errorf("error : %v, output: %s", err, output)
	}
	return nil
}

func ExecuteCommand(command string) (string, error) {
	return executeCommand("/bin/bash", []string{"-c", command})
}

func ExecuteCommands(commands []string) (results []string, err error) {
	for _, command := range commands {
		result, err := ExecuteCommand(command)
		if err != nil {
			results = append(results, "error")
		} else {
			results = append(results, result)
		}
	}
	return results, nil
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
