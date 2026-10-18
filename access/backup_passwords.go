package access

import (
	"crypto/rand"
	"errors"
	"strings"
)

// backupAlphabet leaves out the letters and digits that read alike.
const backupAlphabet = "abcdefghjkmnpqrstuvwxyz23456789"

// BackupPasswords returns count new backup passwords, each three groups
// of four characters, about 59 bits, and their hashes in the same order.
// The passwords are shown once; only the hashes are kept.
func BackupPasswords(count int) ([]string, []string, error) {
	if count < 1 || count > 100 {
		return nil, nil, errors.New("backup passwords come 1 to 100 at a time")
	}
	passwords := make([]string, count)
	hashes := make([]string, count)
	for i := range passwords {
		var b strings.Builder
		for j, c := range randomBytes(12) {
			if j > 0 && j%4 == 0 {
				b.WriteByte('-')
			}
			b.WriteByte(backupAlphabet[int(c)%len(backupAlphabet)])
		}
		hash, err := HashPassword(b.String())
		if err != nil {
			return nil, nil, err
		}
		passwords[i], hashes[i] = b.String(), hash
	}
	return passwords, hashes, nil
}

// randomBytes reads n bytes from the system's random source. Bytes at or
// above the largest multiple of the alphabet are drawn again, so every
// character is equally likely.
func randomBytes(n int) []byte {
	limit := byte(256 - 256%len(backupAlphabet))
	out := make([]byte, 0, n)
	buf := make([]byte, 1)
	for len(out) < n {
		rand.Read(buf)
		if buf[0] < limit {
			out = append(out, buf[0])
		}
	}
	return out
}
