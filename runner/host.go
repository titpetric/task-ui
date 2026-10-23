package runner

import (
	"path/filepath"
	"slices"
	"strings"
)

// serverOnly are the variables of the server's own configuration that a
// run sets for itself: TASK_UI_WORKSPACE names the run's workspace, and
// the server's names its workspace folder.
var serverOnly = []string{"TASK_UI_WORKSPACE=", "TASK_UI_HOST_DATA=", "TASK_UI_HOST_PWD="}

// inherited returns the server's environment as a run starts from it,
// without serverOnly.
func inherited(env []string) []string {
	kept := env[:0:0]
	for _, kv := range env {
		if !slices.ContainsFunc(serverOnly, func(prefix string) bool { return strings.HasPrefix(kv, prefix) }) {
			kept = append(kept, kv)
		}
	}
	return kept
}

// hostPath returns a directory of the run as the host sees it.
func (m *Manager) hostPath(dir string) string {
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	return m.opts.HostPaths.Host(dir)
}
