package app

import (
	"context"
	"path/filepath"
	"time"

	"github.com/titpetric/task-ui/feature/inventory/hostpath"
	"github.com/titpetric/task-ui/integration/beszel"
	"github.com/titpetric/task-ui/integration/ssh/sshprobe"
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

// promptHosts lists the inventory for a prompt's run: the hosts, largest
// first, with who task-ui signs in to each as, waiting at most 10s for
// the hosts not yet probed.
//
// held is the inventory the Inventory page keeps, a name by address. A host
// a person entered is in the list whether the hub reports it or not, and a
// hub that answers nothing leaves the rows.
func promptHosts(hosts *beszel.Inventory, prober *sshprobe.Prober, held func(ctx context.Context) map[string]string) func(ctx context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		systems, err := hosts.Systems(ctx)
		rows := held(ctx)
		if err != nil && len(rows) == 0 {
			return "", err
		}
		systems = beszel.ByCapacity(append(systems, unreported(systems, rows)...))
		names := make([]string, len(systems))
		for i, system := range systems {
			names[i] = system.Host
		}
		return beszel.Hosts(systems, prober.Logins(ctx, names, 10*time.Second)), nil
	}
}
