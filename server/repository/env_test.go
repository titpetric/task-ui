package repository

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/server/model"
)

func TestEnvSecrets(t *testing.T) {
	env := &Env{Root: t.TempDir()}

	if err := env.SetSecret(model.Secret{Scope: ScopeAll, Key: "TOKEN", Value: "shared"}); err != nil {
		t.Fatal(err)
	}
	if err := env.SetSecret(model.Secret{Scope: "infra", Key: "TOKEN", Value: "mine with space"}); err != nil {
		t.Fatal(err)
	}
	if err := env.SetSecret(model.Secret{Scope: "infra", Key: "REGION", Value: "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := env.SetSecret(model.Secret{Scope: "infra", Key: "1BAD", Value: "x"}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("invalid key accepted: %v", err)
	}
	if err := env.SetSecret(model.Secret{Scope: "../etc", Key: "X", Value: "x"}); err == nil {
		t.Fatal("invalid scope accepted")
	}

	info, err := os.Stat(filepath.Join(env.Root, "secrets", "infra.env"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("secrets file mode: %v %v", info, err)
	}

	list := env.Secrets()
	got := []string{}
	for _, s := range list {
		got = append(got, s.Scope+"/"+s.Key)
	}
	if want := []string{"all/TOKEN", "infra/REGION", "infra/TOKEN"}; !slices.Equal(got, want) {
		t.Fatalf("secrets: got %v, want %v", got, want)
	}

	for_ := env.SecretsFor("infra")
	if for_["TOKEN"] != "mine with space" || for_["REGION"] != "eu" {
		t.Fatalf("folder secrets: %v", for_)
	}
	if other := env.SecretsFor("web"); other["TOKEN"] != "shared" || len(other) != 1 {
		t.Fatalf("shared secrets: %v", other)
	}

	if err := env.DeleteSecret("infra", "REGION"); err != nil {
		t.Fatal(err)
	}
	if err := env.DeleteSecret("infra", "REGION"); err == nil {
		t.Fatal("deleted a missing secret")
	}
	if _, ok := env.Secret("infra", "REGION"); ok {
		t.Fatal("secret still there")
	}
}

func TestEnvFiles(t *testing.T) {
	env := &Env{Root: t.TempDir()}
	dir := t.TempDir()
	project := &model.Project{ID: "infra", Dir: dir, EnvFiles: []string{".env", "conf/deploy.env", "../outside.env"}}

	os.WriteFile(filepath.Join(dir, ".env"), []byte("PORT=8080\n# note\nNAME=\"my app\"\n"), 0o644)

	files := env.Files(project)
	if len(files) != 2 {
		t.Fatalf("files: %+v", files)
	}
	if !files[0].Exists || files[0].Managed || !slices.Equal(files[0].Keys, []string{"NAME", "PORT"}) {
		t.Fatalf(".env: %+v", files[0])
	}
	if files[1].Exists || files[1].Managed {
		t.Fatalf("deploy.env: %+v", files[1])
	}

	env.SetSecret(model.Secret{Scope: ScopeAll, Key: "TOKEN", Value: "s3cret"})
	env.SetSecret(model.Secret{Scope: "infra", Key: "PORT", Value: "9090"})

	if err := env.Save(project, ".env", "PORT=8080\r\nNAME=app\r\n"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, ".env"))
	content := string(data)
	if !strings.HasPrefix(content, "PORT=8080\nNAME=app\n\n"+generatedMarker+"\nPORT=9090\nTOKEN=s3cret\n") {
		t.Fatalf("generated .env: %q", content)
	}

	// Reading it back shows the managed content, not the secrets.
	file, _ := env.File(project, ".env")
	if !file.Managed || file.Content != "PORT=8080\nNAME=app\n" {
		t.Fatalf("managed .env: %+v", file)
	}

	// A file never saved and not in the folder is generated from the secrets.
	if err := env.Generate(project); err == nil {
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
	unmanaged := &model.Project{ID: "web", Dir: t.TempDir(), EnvFiles: []string{".env"}}
	os.WriteFile(filepath.Join(unmanaged.Dir, ".env"), []byte("A=1\n\n"+generatedMarker+"\nTOKEN=old\n"), 0o600)
	file, _ = env.File(unmanaged, ".env")
	if file.Content != "A=1\n" || !slices.Equal(file.Keys, []string{"A"}) {
		t.Fatalf("stripped: %+v", file)
	}
}

func TestDotenvQuoting(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.env")
	vars := map[string]string{"A": "plain", "B": "has space", "C": `say "hi"`, "D": "", "E": "a#b"}
	if err := writeDotenv(path, vars); err != nil {
		t.Fatal(err)
	}
	back, err := ReadDotenv(path)
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range vars {
		if back[key] != want {
			t.Errorf("%s: got %q, want %q", key, back[key], want)
		}
	}
}
