package model

import (
	"crypto/rand"
	"encoding/base32"
	"regexp"
	"strings"
	"time"
)

// IDs are 26 characters: ten of the time in milliseconds, so rows sort
// by creation, and sixteen random, so two made in the same instant stay
// apart. The alphabet is Crockford base32, lower case.

var idAlphabet = base32.NewEncoding("0123456789abcdefghjkmnpqrstvwxyz").WithPadding(base32.NoPadding)

var validID = regexp.MustCompile(`^[0-9a-z]{26}$`)

// NewID returns a new identifier for a row.
func NewID() string {
	return NewIDAt(time.Now())
}

// NewIDAt returns an identifier that sorts as of t.
func NewIDAt(t time.Time) string {
	stamp := make([]byte, 8)
	ms := uint64(t.UnixMilli())
	for i := 7; i >= 0; i-- {
		stamp[i] = byte(ms)
		ms >>= 8
	}
	random := make([]byte, 10)
	_, _ = rand.Read(random)
	head := idAlphabet.EncodeToString(stamp[2:]) // 48 bits of time, 10 chars
	tail := idAlphabet.EncodeToString(random)    // 80 bits, 16 chars
	return strings.ToLower(head + tail)
}

// ValidID reports whether id is one NewID could have made.
func ValidID(id string) bool {
	return validID.MatchString(id)
}
