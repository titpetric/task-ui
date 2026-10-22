package beszel

import (
	"strings"
)

// Hosts lists the systems for a run, one a line, in their order: the
// name, the host, the user that signs in to it and up or down, separated
// by tabs. logins holds user@host for each host task-ui signs in to, as
// sshprobe finds them; a host without one has an empty user.
func Hosts(systems []System, logins map[string]string) string {
	lines := make([]string, 0, len(systems))
	for _, system := range systems {
		user, _, _ := strings.Cut(logins[system.Host], "@")
		state := "down"
		if system.Status == "up" {
			state = "up"
		}
		lines = append(lines, strings.Join([]string{system.Name, system.Host, user, state}, "\t"))
	}
	return strings.Join(lines, "\n")
}
