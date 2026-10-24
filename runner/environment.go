package runner

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
)

// EnvironmentSetting names the instance setting that holds how much of
// the server's environment a run starts with: EnvNone, EnvAuto or
// EnvAll; empty is EnvNone.
const EnvironmentSetting = "runs.parent_env"

// The modes of the parent environment.
const (
	// EnvNone passes the variables of BaseEnv alone, less the sensitive
	// ones: the secure default.
	EnvNone = "none"
	// EnvAuto passes every variable but the sensitive ones.
	EnvAuto = "auto"
	// EnvAll passes every variable.
	EnvAll = "all"
)

// EnvModes are the modes, in the order the Instance page lists them.
var EnvModes = []string{EnvNone, EnvAuto, EnvAll}

// ErrEnvironmentInvalid refuses a mode other than EnvModes.
var ErrEnvironmentInvalid = errors.New("the parent environment is none, auto or all")

// BaseEnv are the variables of the server a run gets under EnvNone: a
// name, or a prefix ending in *. The rest of what a run needs task-ui
// sets itself: the run's TASK_UI_* facts, its HOME, SSH_AUTH_SOCK of its
// own agent, the folder's secrets and the integrations' variables.
var BaseEnv = []string{
	"PATH", "HOME", "USER", "LOGNAME", "SHELL", "HOSTNAME", "TMPDIR",
	"LANG", "LANGUAGE", "LC_*", "TZ", "TERM",
	"TASK_UI_*",
	"DOCKER_HOST", "ATKINS_DB", "DOCKER_CONFIG", "GIT_CONFIG_GLOBAL", "XDG_*", "GOPATH", "HISTFILE", "CLAUDE_CONFIG_DIR",
}

// SensitiveWords mark a variable that may hold a credential when its
// name holds one of them, in any case: WORKSPACE_KEY, GITHUB_TOKEN,
// DB_PASSWORD. EnvNone and EnvAuto leave such a variable out.
var SensitiveWords = []string{"KEY", "TOKEN", "SECRET", "PASS", "CREDENTIAL", "PRIVATE"}

// Sensitive reports a name that holds one of SensitiveWords.
func Sensitive(name string) bool {
	upper := strings.ToUpper(name)
	return slices.ContainsFunc(SensitiveWords, func(word string) bool { return strings.Contains(upper, word) })
}

// Base reports a name that BaseEnv lists.
func Base(name string) bool {
	return slices.ContainsFunc(BaseEnv, func(pattern string) bool {
		if prefix, ok := strings.CutSuffix(pattern, "*"); ok {
			return strings.HasPrefix(name, prefix)
		}
		return name == pattern
	})
}

// Forwarded reports whether a run gets a variable of the server under a
// mode: under EnvAll every one, under EnvAuto every one but the
// sensitive ones, else those of BaseEnv but the sensitive ones.
func Forwarded(name, mode string) bool {
	switch mode {
	case EnvAll:
		return true
	case EnvAuto:
		return !Sensitive(name)
	}
	return Base(name) && !Sensitive(name)
}

// ParentEnv returns what of environ a run gets under a mode, in order.
func ParentEnv(environ []string, mode string) []string {
	kept := []string{}
	for _, kv := range inherited(environ) {
		name, _, _ := strings.Cut(kv, "=")
		if Forwarded(name, mode) {
			kept = append(kept, kv)
		}
	}
	return kept
}

// Environment is how much of the server's environment the runs start
// with: the mode an admin sets on the Instance page, read on every
// start, so a change applies to the next run.
type Environment struct {
	Settings SettingValues
	// Environ is the server's environment, nil for os.Environ.
	Environ func() []string
}

// Mode returns the mode in effect: EnvAll for an Environment of no
// settings, EnvNone when unset or unreadable.
func (e *Environment) Mode(ctx context.Context) string {
	if e == nil || e.Settings == nil {
		return EnvAll
	}
	values, err := e.Settings.Values(ctx)
	if err != nil || !slices.Contains(EnvModes, values[EnvironmentSetting]) {
		return EnvNone
	}
	return values[EnvironmentSetting]
}

// Set stores a mode for the next run; it fails on a mode other than
// EnvModes, or without settings.
func (e *Environment) Set(ctx context.Context, mode string) error {
	if !slices.Contains(EnvModes, mode) {
		return ErrEnvironmentInvalid
	}
	if e == nil || e.Settings == nil {
		return errors.New("the instance settings aren't available")
	}
	return e.Settings.Set(ctx, map[string]string{EnvironmentSetting: mode})
}

// Env returns the variables of the server a run starts with now.
func (e *Environment) Env(ctx context.Context) []string {
	return ParentEnv(e.environ(), e.Mode(ctx))
}

// Names returns the server's variables, sorted, as EnvName, never a value.
func (e *Environment) Names() []EnvName {
	names := []EnvName{}
	for _, kv := range inherited(e.environ()) {
		name, _, _ := strings.Cut(kv, "=")
		if name == "" || slices.ContainsFunc(names, func(n EnvName) bool { return n.Name == name }) {
			continue
		}
		names = append(names, EnvName{Name: name, Sensitive: Sensitive(name), None: Forwarded(name, EnvNone), Auto: Forwarded(name, EnvAuto)})
	}
	slices.SortFunc(names, func(a, b EnvName) int { return strings.Compare(a.Name, b.Name) })
	return names
}

func (e *Environment) environ() []string {
	if e != nil && e.Environ != nil {
		return e.Environ()
	}
	return os.Environ()
}
