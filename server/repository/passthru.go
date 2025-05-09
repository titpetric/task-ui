package repository

import (
	"bytes"
	"fmt"
	"os/exec"
)

func passthru(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to execute command: %v", err)
	}

	return output.String(), nil
}
