package envfile

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/folder"
	"github.com/titpetric/task-ui/secret"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	return &Store{Root: t.TempDir(), Secrets: &secret.Store{Root: t.TempDir()}}
}

func TestStore_Files(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{".env", "conf/deploy.env", "../outside.env"}}
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8080\n# note\nNAME=\"my app\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	files := s.Files(f)
	if len(files) != 2 {
		t.Fatalf("files: %+v", files)
	}
	if !files[0].Exists || files[0].Managed || !slices.Equal(files[0].Keys, []string{"NAME", "PORT"}) {
		t.Fatalf(".env: %+v", files[0])
	}
	if files[1].Exists || files[1].Managed {
		t.Fatalf("deploy.env: %+v", files[1])
	}
	if _, err := s.File(f, "../outside.env"); err == nil {
		t.Fatal("path outside the folder accepted")
	}
	if _, err := s.File(f, "$HOME/.env"); err == nil {
		t.Fatal("path with a variable accepted")
	}
}

func TestStore_Save(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{".env"}}
	if err := s.Secrets.Set(secret.Secret{Scope: secret.ScopeAll, Key: "TOKEN", Value: "s3cret"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Secrets.Set(secret.Secret{Scope: "infra", Key: "PORT", Value: "9090"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Save(f, File{Path: ".env", Content: "PORT=8080\r\nNAME=app\r\n"}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".env"))
	if content := string(data); !strings.HasPrefix(content, "PORT=8080\nNAME=app\n\n"+marker+"\nPORT=9090\nTOKEN=s3cret\n") {
		t.Fatalf("generated .env: %q", content)
	}

	// Reading it back shows the managed content, not the secrets.
	file, _ := s.File(f, ".env")
	if !file.Managed || file.Content != "PORT=8080\nNAME=app\n" {
		t.Fatalf("managed .env: %+v", file)
	}
	if err := s.Save(f, File{Path: "../x.env", Content: "A=1"}); err == nil {
		t.Fatal("saved outside the folder")
	}
}

func TestStore_Generate(t *testing.T) {
	s := newStore(t)
	dir := t.TempDir()
	f := &folder.Folder{ID: "infra", Dir: dir, EnvFiles: []string{"conf/deploy.env", "../outside.env"}}
	if err := s.Secrets.Set(secret.Secret{Scope: secret.ScopeAll, Key: "TOKEN", Value: "s3cret"}); err != nil {
		t.Fatal(err)
	}

	// A file never saved and not in the folder is generated from the secrets.
	if err := s.Generate(f); err == nil {
		t.Fatal("expected the outside path to fail")
	}
	data, err := os.ReadFile(filepath.Join(dir, "conf", "deploy.env"))
	if err != nil || !strings.Contains(string(data), "TOKEN=s3cret") {
		t.Fatalf("generated deploy.env: %q %v", data, err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(dir), "outside.env")); err == nil {
		t.Fatal("wrote outside the folder")
	}

	// A folder file that was generated before reads back without the secrets.
	unmanaged := &folder.Folder{ID: "web", Dir: t.TempDir(), EnvFiles: []string{".env"}}
	if err := os.WriteFile(filepath.Join(unmanaged.Dir, ".env"), []byte("A=1\n\n"+marker+"\nTOKEN=old\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	file, _ := s.File(unmanaged, ".env")
	if file.Content != "A=1\n" || !slices.Equal(file.Keys, []string{"A"}) {
		t.Fatalf("stripped: %+v", file)
	}

	// Without secrets the file is the content alone.
	plain := newStore(t)
	if err := plain.Generate(unmanaged); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(filepath.Join(unmanaged.Dir, ".env"))
	if string(data) != "A=1\n" {
		t.Fatalf("plain generate: %q", data)
	}
}
