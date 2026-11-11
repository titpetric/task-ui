package hostinfo

import (
	"strings"
	"time"
)

// Facts are what a host told over ssh: its durable filesystems, and its
// containers with their sizes when docker answered. At is when it told.
type Facts struct {
	Filesystems []Filesystem
	Containers  []DockerContainer
	Docker      bool
	At          time.Time
	Err         error
}

// script is what the host runs: df in bytes, then docker ps with sizes
// as JSON lines, as the login or else through sudo without a password,
// then whether docker answered.
const script = `df -PT -B1 2>/dev/null; echo '@@docker'; ` +
	`{ docker ps --size --format '{{json .}}' 2>/dev/null || sudo -n docker ps --size --format '{{json .}}' 2>/dev/null; }; echo "@@exit $?"`

// ParseFacts reads the output of script.
func ParseFacts(out string) *Facts {
	df, rest, _ := strings.Cut(out, "@@docker\n")
	ps, exit, _ := strings.Cut(rest, "@@exit ")
	return &Facts{
		Filesystems: Durable(ParseDF(df)),
		Containers:  ParseDockerPS(ps),
		Docker:      strings.TrimSpace(exit) == "0",
	}
}
