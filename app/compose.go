package app

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ComposeProject returns the docker compose project of the container
// task-ui runs in: the com.docker.compose.project label of the container
// named by the host name, as docker inspects it. "" outside a container,
// without docker, or for a container compose did not start.
func ComposeProject(docker string) string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, docker, "inspect", "--format", `{{ index .Config.Labels "com.docker.compose.project" }}`, host)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	project := strings.TrimSpace(string(out))
	if project == "<no value>" {
		return ""
	}
	return project
}
