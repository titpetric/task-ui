package runner

import (
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/layout"
)

// homeKept are the variables a run gets set to the server's HOME, so what
// the instance keeps there stays in use when HOME moves to the
// workspace's home: the registry logins, Claude Code's settings and
// sessions, the shared build caches of the XDG tools and Go's module
// cache. One the server sets already keeps its value. Everything else a
// tool reads from HOME, the git identity, the shell's history, the XDG
// configuration, data and state, is the workspace's own, in its home.
var homeKept = []struct{ name, path string }{
	{"DOCKER_CONFIG", ".docker"},
	{"CLAUDE_CONFIG_DIR", ".claude"},
	{"XDG_CACHE_HOME", ".cache"},
	{"GOPATH", "go"},
}

// HomeEnv returns HOME, TASK_UI_HOME, ATKINS_DB and the kept variables
// for a run with HOME at home, from environ; nil for no home.
func HomeEnv(environ []string, home string) []string {
	if home == "" {
		return nil
	}
	server := lookup(environ, "HOME")
	env := []string{}
	if server != "" {
		for _, kept := range homeKept {
			if lookup(environ, kept.name) == "" {
				env = append(env, kept.name+"="+filepath.Join(server, filepath.FromSlash(kept.path)))
			}
		}
	}
	return append(env, "HOME="+home, "TASK_UI_HOME="+home, "ATKINS_DB="+layout.AtkinsDB(home, server))
}

// lookup returns a variable of environ, the last one set winning, as it
// does for a process.
func lookup(environ []string, name string) string {
	value := ""
	for _, kv := range environ {
		if k, v, ok := strings.Cut(kv, "="); ok && k == name {
			value = v
		}
	}
	return value
}
