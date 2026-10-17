package folder

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/titpetric/task-ui/model"
)

func TestConfig_Apply(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\n")
	f := Load("tools", dir)
	Config{Title: "Tools", Description: "The tools", Allow: []string{"ops"}, Command: "gmake", Jobs: map[string][]string{"build": {"admin"}}}.Apply(f)
	if f.Title != "Tools" || f.Description != "The tools" || !slices.Equal(f.Allow, []string{"ops"}) {
		t.Fatalf("overrides: %+v", f)
	}
	if !slices.Equal(f.Sources[0].Command, []string{"gmake"}) || !slices.Equal(f.Job("build").Command, []string{"gmake", "build"}) {
		t.Fatalf("command override: %v %v", f.Sources[0].Command, f.Job("build").Command)
	}
	if !slices.Equal(f.Job("build").Allow, []string{"admin"}) {
		t.Fatalf("job allow: %+v", f.Job("build"))
	}

	write(t, filepath.Join(dir, "package.json"), `{"scripts": {"build": "vite build"}}`)
	f = Load("tools", dir)
	Config{Command: "gmake", Commands: map[string]string{"npm": "pnpm run"}}.Apply(f)
	if !slices.Equal(f.Job("make:build").Command, []string{"make", "build"}) {
		t.Fatalf("command applies to a single source only: %v", f.Job("make:build").Command)
	}
	if !slices.Equal(f.Job("npm:build").Command, []string{"pnpm", "run", "build"}) {
		t.Fatalf("per kind command: %v", f.Job("npm:build").Command)
	}
}

func TestConfigOf(t *testing.T) {
	t.Parallel()
	project := &model.Project{
		Folder:      "infra",
		Title:       "Infrastructure",
		Description: "Deploys",
		Allow:       model.EncodeList([]string{"ops"}),
		Commands:    model.EncodeStrings(map[string]string{"make": "gmake"}),
		Sources:     model.EncodeList([]string{"make"}),
		JobAllow:    model.EncodeMap(map[string][]string{"build": {"admin"}}),
	}
	cfg := ConfigOf(project)
	if cfg.ID != "infra" || cfg.Title != "Infrastructure" || cfg.Description != "Deploys" {
		t.Fatalf("identity: %+v", cfg)
	}
	if !slices.Equal(cfg.Allow, []string{"ops"}) || cfg.Commands["make"] != "gmake" || !slices.Equal(cfg.Sources, []string{"make"}) || !slices.Equal(cfg.Jobs["build"], []string{"admin"}) {
		t.Fatalf("lists: %+v", cfg)
	}
	if empty := ConfigOf(nil); empty.ID != "" {
		t.Fatalf("nil project: %+v", empty)
	}
}

func TestConfigsOf(t *testing.T) {
	t.Parallel()
	configs := ConfigsOf([]model.Project{{Folder: "a"}, {Folder: "b"}})
	if len(configs) != 2 || configs[1].ID != "b" {
		t.Fatalf("configs: %+v", configs)
	}
}

func TestConfig_ApplyHours(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "Makefile"), "build:\n\tgo build\ntest:\n\tgo test\n")
	f := Load("tools", dir)
	Config{Hours: "mon-fri 9-17", JobHours: map[string]string{"test": "daily 0-24", "nothing": "sun 1-2"}}.Apply(f)
	if f.Hours != "mon-fri 9-17" || f.Job("test").Hours != "daily 0-24" || f.Job("build").Hours != "" {
		t.Fatalf("hours: %+v", f)
	}
	if f.HoursOf(f.Job("build")) != "mon-fri 9-17" || f.HoursOf(f.Job("test")) != "daily 0-24" || f.HoursOf(nil) != "mon-fri 9-17" {
		t.Fatalf("hours of: %q %q", f.HoursOf(f.Job("build")), f.HoursOf(f.Job("test")))
	}
	if (*Folder)(nil).HoursOf(nil) != "" {
		t.Fatal("nil folder has hours")
	}

	cfg := ConfigOf(&model.Project{Folder: "tools", Hours: "sat 1-2", JobHours: `{"build":"sun 1-2"}`})
	if cfg.Hours != "sat 1-2" || cfg.JobHours["build"] != "sun 1-2" {
		t.Fatalf("config of: %+v", cfg)
	}
}
