package view_test

import (
	"io/fs"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestAssetsFS(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "/static/", view.AssetPrefix)
	for _, name := range []string{
		"task-ui.css", "task-ui.js", "icon.svg", "vendor/xterm.js", "vendor/addon-fit.js", "vendor/xterm.css",
		"fonts/Inter-Regular.woff2", "fonts/Inter-Medium.woff2", "fonts/Inter-SemiBold.woff2", "fonts/Inter.LICENSE",
		"fonts/JetBrainsMono-Regular.woff2", "fonts/JetBrainsMono-Medium.woff2", "fonts/JetBrainsMono-SemiBold.woff2", "fonts/JetBrainsMono.LICENSE",
	} {
		data, err := fs.ReadFile(view.AssetsFS(), name)
		assert.NoError(t, err, name)
		assert.NotEmpty(t, data, name)
	}
}
