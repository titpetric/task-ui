package access

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(hash, "$2") {
		t.Fatalf("not a bcrypt hash: %q", hash)
	}
	user := &User{Name: "a", Password: hash}
	if !CheckPassword(user, "secret") {
		t.Fatal("right password refused")
	}
	if CheckPassword(user, "wrong") {
		t.Fatal("wrong password accepted")
	}
}

func TestCheckPassword(t *testing.T) {
	if CheckPassword(nil, "anything") {
		t.Fatal("nil user matched")
	}
	if CheckPassword(&User{Password: "not a hash"}, "anything") {
		t.Fatal("garbage hash matched")
	}
}
