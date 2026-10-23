// Package hostpath maps the paths of the container task-ui runs in to
// the paths of its host: a job's `docker run -v $PWD:/src` goes to the
// host's docker daemon, which resolves the path on the host.
package hostpath

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Mounts are the bind mounts of a container, the deepest mount point
// first: each a directory of the host at a path of the container.
type Mounts []mount

// mount is a host directory, Source, at Target in the container.
type mount struct {
	Source, Target string
}

// Read returns the mounts of the process's container, from
// /proc/self/mountinfo; nil outside Linux or a container.
func Read() Mounts {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return nil
	}
	defer f.Close()
	mounts, _ := Parse(f)
	return mounts
}

// Parse reads mountinfo lines: a bind mount of a directory shows its
// path within its file system, the host path for a file system mounted
// at the host's /, at its mount point. The root mount and the mounts of
// a file system's own root, such as /proc, map nothing.
func Parse(r io.Reader) (Mounts, error) {
	var mounts Mounts
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 5 {
			continue
		}
		source, target := unescape(fields[3]), unescape(fields[4])
		if source == "/" || target == "/" || !strings.HasPrefix(source, "/") {
			continue
		}
		mounts = append(mounts, mount{Source: source, Target: target})
	}
	sort.SliceStable(mounts, func(i, j int) bool { return len(mounts[i].Target) > len(mounts[j].Target) })
	return mounts, scanner.Err()
}

// Host returns the host path of a path of the container: under the
// deepest mount that holds it, its source with the rest of the path;
// the path itself when no mount holds it, as outside a container.
func (m Mounts) Host(path string) string {
	path = filepath.Clean(path)
	for _, mount := range m {
		if path == mount.Target {
			return mount.Source
		}
		if rest, ok := strings.CutPrefix(path, mount.Target+"/"); ok {
			return filepath.Join(mount.Source, rest)
		}
	}
	return path
}

// unescape undoes the octal escapes of mountinfo, \040 for a space.
func unescape(field string) string {
	if !strings.Contains(field, `\`) {
		return field
	}
	var b strings.Builder
	for i := 0; i < len(field); i++ {
		if field[i] == '\\' && i+3 < len(field) {
			if n, ok := octal(field[i+1 : i+4]); ok {
				b.WriteByte(n)
				i += 3
				continue
			}
		}
		b.WriteByte(field[i])
	}
	return b.String()
}

// octal reads three octal digits as a byte.
func octal(digits string) (byte, bool) {
	var n byte
	for _, d := range []byte(digits) {
		if d < '0' || d > '7' {
			return 0, false
		}
		n = n*8 + d - '0'
	}
	return n, true
}
