package beszel

import (
	"strings"
)

// Hosts lists the systems a line each, tab separated: name, host, the user of its login, and up or down.
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
