package app_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// TASK_UI_HOST_DATA names the data directory on the host, which the
// Instance page shows; without it, the page shows what the mounts say.
func TestNew_HostData(t *testing.T) {
	t.Parallel()
	for host, want := range map[string]string{"/mnt/apps/task/data": "/mnt/apps/task/data", "": "Data on the host"} {
		root := t.TempDir()
		cfg := &app.Config{
			Listen: "127.0.0.1:0", Root: root, DataDir: filepath.Join(t.TempDir(), "data"), DB: ":memory:", SSHDir: t.TempDir(),
			SessionTTL: time.Hour, Keep: time.Minute, Grace: time.Second, AutoLogin: "admin", HostData: host,
		}
		assert.NoError(t, cfg.Validate())
		a, err := app.New(context.Background(), cfg, "test")
		assert.NoError(t, err)
		status, body, _ := get(t, a.Handler(), "/settings/instance")
		assert.Equal(t, http.StatusOK, status)
		assert.Contains(t, body, want)
		assert.Contains(t, body, "<code>TASK_UI_HOST_DATA</code>")
		a.Close()
	}
}
