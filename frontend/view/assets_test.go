package view_test

import (
	"io/fs"
	"regexp"
	"slices"
	"strings"
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

// The version of the assets is a short hash of them, the same every call.
func TestAssetVersion(t *testing.T) {
	t.Parallel()
	v := view.AssetVersion()
	assert.Len(t, v, 12)
	assert.Equal(t, v, view.AssetVersion())
}

// The stylesheet imports every file in css/ once, each with the version
// of the assets, and nothing in css/ goes without it.
func TestStylesheet(t *testing.T) {
	t.Parallel()

	css := string(view.Stylesheet())
	assert.Equal(t, css, string(view.Stylesheet()))

	files, err := fs.Glob(view.AssetsFS(), "css/*.css")
	assert.NoError(t, err)
	assert.NotEmpty(t, files)
	for _, name := range files {
		assert.Equal(t, 1, strings.Count(css, `"`+name+`?v=`+view.AssetVersion()+`"`), name)
	}

	imports := regexp.MustCompile(`@import url\("([^"]+)"\);`).FindAllStringSubmatch(css, -1)
	assert.Len(t, imports, len(files))
	// The order is the cascade: the tokens first, the controls before the
	// forms that place them, and the phone's folds last, over every page.
	order := []string{}
	for _, m := range imports {
		order = append(order, strings.TrimSuffix(m[1], "?v="+view.AssetVersion()))
	}
	assert.Equal(t, "css/tokens.css", order[0])
	assert.Equal(t, "css/mobile.css", order[len(order)-1])
	assert.Less(t, slices.Index(order, "css/controls.css"), slices.Index(order, "css/forms.css"))
	assert.Less(t, slices.Index(order, "css/forms.css"), slices.Index(order, "css/filters.css"))
	for _, m := range imports {
		assert.True(t, strings.HasSuffix(m[1], "?v="+view.AssetVersion()), m[1])
		data, err := fs.ReadFile(view.AssetsFS(), strings.TrimSuffix(m[1], "?v="+view.AssetVersion()))
		assert.NoError(t, err, m[1])
		assert.NotEmpty(t, data, m[1])
	}
	assert.NotContains(t, css, `css/forms.css"`)
}
