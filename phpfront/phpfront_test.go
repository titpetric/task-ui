package phpfront_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

// The embedded tree holds the application: the autoloader, the domain, the
// theme and the tools. A release that shipped without one of them would answer
// its first request with a missing file.
func TestApp(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"composer.json", "composer.lock", "ext/autoload.php", "phpscript.yml",
		"code/Module.php", "code/Task/Row.php", "code/Task/Model/Workspace.php",
		"stubs/autoload.php", "stubs/host.php",
		"tools/models.php", "tools/templates/model.tpl",
		"theme/assets/task-ui.css", "theme/assets/task-ui.js",
		"ext/titpetric/mdedit/dist/mdedit.global.js",
	} {
		data, err := fs.ReadFile(phpfront.App(), name)
		assert.NoError(t, err, name)
		assert.NotEmpty(t, data, name)
	}
}

// composer installs into ext, because x/mod/zip drops a file more than one
// level under a directory named vendor and the autoload tree is three deep: a
// release would ship vendor/autoload.php with nothing left for it to load.
func TestApp_InstallsIntoExt(t *testing.T) {
	t.Parallel()

	_, err := fs.ReadFile(phpfront.App(), "ext/composer/autoload_real.php")
	assert.NoError(t, err)
	_, err = fs.Stat(phpfront.App(), "vendor")
	assert.Error(t, err)
}

// TestApp_ThemeBundlesInExt holds the browser bundles in ext and no vendor directory.
func TestApp_ThemeBundlesInExt(t *testing.T) {
	t.Parallel()

	for _, name := range []string{
		"theme/assets/ext/addon-fit.js",
		"ext/titpetric/mdedit/LICENSE",
		"ext/titpetric/mdedit/dist/mdedit.global.js",
		"ext/titpetric/mdedit/dist/mdedit.min.css",
		"theme/assets/ext/xterm.LICENSE",
		"theme/assets/ext/xterm.css",
		"theme/assets/ext/xterm.js",
	} {
		_, err := fs.Stat(phpfront.App(), name)
		assert.NoError(t, err, name)
	}
	_, err := fs.Stat(phpfront.App(), "theme/assets/vendor")
	assert.Error(t, err)
}

// The theme's assets come out of the embed, so /static/ serves one copy of the
// stylesheet while the Go pages are still there.
func TestThemeAssets(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(phpfront.ThemeAssets(), "task-ui.css")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "@import url(")
}

// The entry stylesheet and the sheets beside it name each other, both ways. A
// sheet renamed without its @import answers 404 and the page loses every rule
// it held; a sheet added without one is never fetched at all. Neither shows up
// in a build or a render, so the pairing is a test.
//
// internal/assets.importURL is the pattern that versions these at serve time,
// so an @import this does not match is one the browser gets without a ?v= and
// keeps for the hour MaxAge allows.
func TestStylesheetImports(t *testing.T) {
	t.Parallel()

	theme := phpfront.ThemeAssets()
	data, err := fs.ReadFile(theme, "task-ui.css")
	assert.NoError(t, err)

	imported := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		_, rest, found := strings.Cut(line, `@import url("`)
		if !found {
			continue
		}
		name, _, _ := strings.Cut(rest, `"`)
		assert.True(t, strings.HasPrefix(name, "css/"), name+": outside css/, so it is served unversioned")
		_, err := fs.ReadFile(theme, name)
		assert.NoError(t, err, name+": imported by task-ui.css and not in the theme")
		imported[name] = true
	}
	assert.True(t, len(imported) > 1, "the entry stylesheet imports the sheets")

	sheets, err := fs.Glob(theme, "css/*.css")
	assert.NoError(t, err)
	for _, name := range sheets {
		assert.True(t, imported[name], name+": in the theme and imported by nothing")
	}
}

// The head links four files outside the @import graph, the ext sheets and
// the voice window's among them, and a rename of one of those 404s with no
// log either. The href of every link and script of the chrome is read off the
// templates and stat-ed in the theme.
func TestStylesheetLinks(t *testing.T) {
	t.Parallel()

	theme := phpfront.ThemeAssets()
	found := 0
	for _, tpl := range []string{"theme/templates/components/head.tpl", "theme/templates/components/footer.tpl", "theme/templates/index.tpl"} {
		data, err := fs.ReadFile(phpfront.App(), tpl)
		assert.NoError(t, err, tpl)
		for _, ref := range themeRefs(string(data)) {
			found++
			_, err := fs.ReadFile(theme, ref)
			assert.NoError(t, err, tpl+": "+ref+" is linked and not in the theme")
		}
	}
	assert.True(t, found >= 4, "the chrome links the sheets and the scripts of the theme")
}

// The icon library is the sheet and the files together: a class naming a file
// that is not there draws nothing, and a file no class names ships for nothing.
// Neither shows up in a build, so the pairing is a test.
func TestIconLibrary(t *testing.T) {
	t.Parallel()

	theme := phpfront.ThemeAssets()
	data, err := fs.ReadFile(theme, "css/icons.css")
	assert.NoError(t, err)

	drawn := map[string]bool{}
	for _, line := range strings.Split(string(data), "\n") {
		_, rest, found := strings.Cut(line, `--icon: url("../icons/`)
		if !found {
			continue
		}
		name, _, _ := strings.Cut(rest, `")`)
		drawn[name] = true
		_, err := fs.Stat(theme, "icons/"+name)
		assert.NoError(t, err, "icons.css draws icons/"+name+" and the theme does not ship it")
	}
	assert.True(t, len(drawn) > 1, "icons.css draws the library")

	files, err := fs.Glob(theme, "icons/*.svg")
	assert.NoError(t, err)
	for _, path := range files {
		_, name, _ := strings.Cut(path, "/")
		assert.True(t, drawn[name], path+" is in the theme and no class draws it")
	}
}

// themeRefs are the theme's own files a template names, as paths under the
// assets: {data.assets}/theme/<path>, the ?v= and the query dropped.
func themeRefs(tpl string) []string {
	var out []string
	for _, part := range strings.Split(tpl, "{data.assets}/theme/")[1:] {
		ref, _, _ := strings.Cut(part, `"`)
		ref, _, _ = strings.Cut(ref, "?")
		if ref != "" && !strings.Contains(ref, "{") {
			out = append(out, ref)
		}
	}
	return out
}

// Extract writes the tree where a runtime can read and write it. Setenv and
// t.Parallel do not mix, so these three run in order.
func TestExtract(t *testing.T) {
	t.Setenv(phpfront.Env, "")

	root, err := phpfront.Extract(t.TempDir())
	assert.NoError(t, err)
	assert.Equal(t, phpfront.Dir, filepath.Base(root))
	for _, name := range []string{"composer.json", "ext/autoload.php", "code/Task/Row.php"} {
		_, err := os.Stat(filepath.Join(root, name))
		assert.NoError(t, err, name)
	}
}

// A file the tree no longer ships is deleted, and the directory it was alone
// in goes with it: an upgrade that drops a module drops its routes.
func TestExtract_Prunes(t *testing.T) {
	t.Setenv(phpfront.Env, "")

	dir := t.TempDir()
	root, err := phpfront.Extract(dir)
	assert.NoError(t, err)

	stale := filepath.Join(root, "modules", "gone", "mod_gone.php")
	assert.NoError(t, os.MkdirAll(filepath.Dir(stale), 0o755))
	assert.NoError(t, os.WriteFile(stale, []byte("<?php"), 0o644))

	_, err = phpfront.Extract(dir)
	assert.NoError(t, err)
	_, err = os.Stat(stale)
	assert.True(t, os.IsNotExist(err), "the file the tree does not ship")
	_, err = os.Stat(filepath.Dir(stale))
	assert.True(t, os.IsNotExist(err), "the directory it was alone in")

	// A shipped file is still there, so the prune reads the tree and not the
	// disk alone.
	_, err = os.Stat(filepath.Join(root, "code", "Task", "Row.php"))
	assert.NoError(t, err)
}

// The checkout override serves a tree in place, so editing a template takes
// effect without a rebuild.
func TestExtract_Checkout(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(phpfront.Env, dir)

	root, err := phpfront.Extract(t.TempDir())
	assert.NoError(t, err)
	assert.Equal(t, dir, root)
}
