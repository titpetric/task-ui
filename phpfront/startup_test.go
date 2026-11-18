package phpfront

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	chi "github.com/go-chi/chi/v5"
)

// TestFront_Bound: a routes.php binds the paths of its tree on the service
// router and the mount answers them beside the annotated endpoints.
func TestFront_Bound(t *testing.T) {
	front := testFront(t, "/development")
	router := chi.NewRouter()
	router.Route("/development", func(r chi.Router) {
		if err := front.Mount(context.Background(), r); err != nil {
			t.Fatalf("mounting: %v", err)
		}
	})

	for _, tc := range []struct {
		name string
		path string
		want string
	}{
		{"a bound path", "/development/bound", "bound GET"},
		{"a path value", "/development/bound/01HZ", "id=01HZ"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if w.Code != http.StatusOK {
				t.Fatalf("status %d: %s", w.Code, w.Body.String())
			}
			if got := w.Body.String(); got != tc.want {
				t.Errorf("want %q, got %q", tc.want, got)
			}
		})
	}
}

// A module's startup job reaches Task::reindex, which starts the search once.
func TestStartup_Reindex(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, ModulesDir, "probe")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	modinfo := "<?php\n// @startup\n$m = new Task\\Module;\n$m->setName(\"taskui/probe\");\nTask::register($m);\n"
	if err := os.WriteFile(filepath.Join(dir, ModInfo), []byte(modinfo), 0o644); err != nil {
		t.Fatal(err)
	}
	boot := "<?php\n// @startup\nTask::reindex();\n"
	if err := os.WriteFile(filepath.Join(dir, "boot.php"), []byte(boot), 0o644); err != nil {
		t.Fatal(err)
	}

	started := 0
	opts := NewOptions(root)
	opts.Search = func(context.Context) error { started++; return nil }
	front, err := New(opts)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	if err := front.Mount(context.Background(), chi.NewRouter()); err != nil {
		t.Fatalf("mounting: %v", err)
	}
	if started != 1 {
		t.Errorf("want the index started once, got %d", started)
	}
}

// TestRegisterModule: a modinfo.php registers through the startup runtime, and
// one that names itself wrongly fails the mount rather than an instance.
func TestRegisterModule(t *testing.T) {
	for _, tc := range []struct {
		name  string
		php   string
		fails string
	}{
		{"a name", "$m->setName(\"taskui/probe\");\n", ""},
		{"no name", "", "a module name is vendor/plugin"},
		{"one half", "$m->setName(\"taskui\");\n", "a module name is vendor/plugin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			dir := filepath.Join(root, ModulesDir, "probe")
			if err := os.MkdirAll(dir, 0o755); err != nil {
				t.Fatal(err)
			}
			source := "<?php\n// @startup\n$m = new Task\\Module;\n" + tc.php + "Task::register($m);\n"
			if err := os.WriteFile(filepath.Join(dir, ModInfo), []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}

			front, err := New(NewOptions(root))
			if err != nil {
				t.Fatalf("new: %v", err)
			}
			err = front.Mount(context.Background(), chi.NewRouter())
			if tc.fails == "" {
				if err != nil {
					t.Fatalf("mounting: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.fails) {
				t.Errorf("want %q, got %v", tc.fails, err)
			}
		})
	}
}
