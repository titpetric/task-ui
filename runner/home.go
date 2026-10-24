package runner

import (
	"path/filepath"
	"strings"
)

// homeKept are the variables a run in a home project gets set to the
// server's HOME, so what the server keeps there stays in use when HOME
// moves: atkins' stats, the registry logins, the git identity, the
// caches and configuration of the XDG tools, Go's module cache, the
// shell's history and Claude Code's settings and sessions. One the
// server sets already keeps its value.
var homeKept = []struct{ name, path string }{
	{"ATKINS_DB", ".atkins/atkins.db"},
	{"DOCKER_CONFIG", ".docker"},
	{"GIT_CONFIG_GLOBAL", ".gitconfig"},
	{"XDG_CONFIG_HOME", ".config"},
	{"XDG_CACHE_HOME", ".cache"},
	{"XDG_DATA_HOME", ".local/share"},
	{"XDG_STATE_HOME", ".local/state"},
	{"GOPATH", "go"},
	{"HISTFILE", ".bash_history"},
	{"CLAUDE_CONFIG_DIR", ".claude"},
}

// HomeEnv returns HOME and TASK_UI_HOME at home, with homeKept under the server's HOME from environ.
// Nil for no home, or a home that is the server's HOME.
func HomeEnv(environ []string, home string) []string {
	server := lookup(environ, "HOME")
	if home == "" || filepath.Clean(home) == filepath.Clean(server) {
		return nil
	}
	env := []string{}
	if server != "" {
		for _, kept := range homeKept {
			if lookup(environ, kept.name) == "" {
				env = append(env, kept.name+"="+filepath.Join(server, filepath.FromSlash(kept.path)))
			}
		}
	}
	return append(env, "HOME="+home, "TASK_UI_HOME="+home)
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
