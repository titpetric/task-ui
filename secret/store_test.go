package secret

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestStore(t *testing.T) {
	s := &Store{Root: t.TempDir()}

	if err := s.Set(Secret{Scope: ScopeAll, Key: "TOKEN", Value: "shared"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(Secret{Scope: "infra", Key: "TOKEN", Value: "mine with space"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(Secret{Scope: "infra", Key: "REGION", Value: "eu"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(Secret{Scope: "infra", Key: "1BAD", Value: "x"}); !errors.Is(err, ErrInvalidKey) {
		t.Fatalf("invalid key accepted: %v", err)
	}
	if err := s.Set(Secret{Scope: "../etc", Key: "X", Value: "x"}); err == nil {
		t.Fatal("invalid scope accepted")
	}

	info, err := os.Stat(filepath.Join(s.Root, "infra.env"))
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("secrets file mode: %v %v", info, err)
	}

	got := []string{}
	for _, item := range s.List() {
		got = append(got, item.Scope+"/"+item.Key)
	}
	if want := []string{"all/TOKEN", "infra/REGION", "infra/TOKEN"}; !slices.Equal(got, want) {
		t.Fatalf("secrets: got %v, want %v", got, want)
	}

	if item, ok := s.Get(Secret{Scope: "infra", Key: "TOKEN"}); !ok || item.Value != "mine with space" {
		t.Fatalf("get: %+v %v", item, ok)
	}
	if _, ok := s.Get(Secret{Scope: "../x", Key: "TOKEN"}); ok {
		t.Fatal("invalid scope got")
	}

	mine := s.For("infra")
	if mine["TOKEN"] != "mine with space" || mine["REGION"] != "eu" {
		t.Fatalf("folder secrets: %v", mine)
	}
	if other := s.For("web"); other["TOKEN"] != "shared" || len(other) != 1 {
		t.Fatalf("shared secrets: %v", other)
	}
	if scoped := s.Scoped("infra"); len(scoped) != 3 {
		t.Fatalf("scoped: %+v", scoped)
	}
	if scoped := s.Scoped("web"); len(scoped) != 1 {
		t.Fatalf("scoped web: %+v", scoped)
	}

	if err := s.Delete(Secret{Scope: "infra", Key: "REGION"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(Secret{Scope: "infra", Key: "REGION"}); err == nil {
		t.Fatal("deleted a missing secret")
	}
	if err := s.Delete(Secret{Scope: "../x", Key: "REGION"}); err == nil {
		t.Fatal("deleted from an invalid scope")
	}
	if _, ok := s.Get(Secret{Scope: "infra", Key: "REGION"}); ok {
		t.Fatal("secret still there")
	}
	if err := s.Delete(Secret{Scope: "infra", Key: "TOKEN"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.Root, "infra.env")); err == nil {
		t.Fatal("empty scope file kept")
	}
}
