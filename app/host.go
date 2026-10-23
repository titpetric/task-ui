package app

import (
	"path/filepath"

	"github.com/titpetric/task-ui/hostpath"
)

// hostPaths returns the mounts of the container task-ui runs in and the
// data directory on the host: TASK_UI_HOST_DATA, or the mounts' answer.
func hostPaths(cfg *Config) (hostpath.Mounts, string) {
	mounts := hostpath.Read()
	if cfg.HostData != "" {
		return mounts, cfg.HostData
	}
	abs, err := filepath.Abs(cfg.DataDir)
	if err != nil {
		return mounts, ""
	}
	return mounts, mounts.Host(abs)
}

// hostDataShown is the data directory on the host for the Instance
// page, empty for the data directory itself.
func hostDataShown(dataDir, host string) string {
	if abs, err := filepath.Abs(dataDir); err == nil && abs == host {
		return ""
	}
	return host
}
