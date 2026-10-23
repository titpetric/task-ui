package hostinfo

import (
	"encoding/json"
	"strconv"
	"strings"
)

// DockerContainer is a container as docker ps --size lists it: its name,
// image, state and status, and the size of its writable layer and of
// the whole, the image under it included, in bytes.
type DockerContainer struct {
	Name        string `json:"name"`
	Image       string `json:"image"`
	State       string `json:"state"`
	Status      string `json:"status"`
	SizeRw      uint64 `json:"size_rw"`
	SizeVirtual uint64 `json:"size_virtual"`
}

// psLine is a line of docker ps --format '{{json .}}'.
type psLine struct {
	Names  string `json:"Names"`
	Image  string `json:"Image"`
	State  string `json:"State"`
	Status string `json:"Status"`
	Size   string `json:"Size"`
}

// ParseDockerPS reads the output of docker ps --size --format
// '{{json .}}', a container a line; it skips a line that is no JSON.
func ParseDockerPS(out string) []DockerContainer {
	var containers []DockerContainer
	for _, line := range strings.Split(out, "\n") {
		var ps psLine
		if json.Unmarshal([]byte(strings.TrimSpace(line)), &ps) != nil || ps.Names == "" {
			continue
		}
		name, _, _ := strings.Cut(ps.Names, ",")
		rw, virtual := parseSize(ps.Size)
		containers = append(containers, DockerContainer{Name: name, Image: ps.Image, State: ps.State, Status: ps.Status, SizeRw: rw, SizeVirtual: virtual})
	}
	return containers
}

// parseSize reads docker's size column, "12.3kB (virtual 150MB)": the
// writable layer and the whole.
func parseSize(size string) (rw, virtual uint64) {
	first, rest, _ := strings.Cut(size, " (virtual ")
	return humanBytes(first), humanBytes(strings.TrimSuffix(rest, ")"))
}

// units are docker's size suffixes, decimal, largest first.
var units = []struct {
	suffix string
	scale  float64
}{{"TB", 1e12}, {"GB", 1e9}, {"MB", 1e6}, {"kB", 1e3}, {"B", 1}}

// humanBytes reads a size as docker prints one, "1.5GB", in bytes; 0
// for what it can't read.
func humanBytes(s string) uint64 {
	s = strings.TrimSpace(s)
	for _, unit := range units {
		if number, ok := strings.CutSuffix(s, unit.suffix); ok {
			value, err := strconv.ParseFloat(strings.TrimSpace(number), 64)
			if err != nil {
				return 0
			}
			return uint64(value * unit.scale)
		}
	}
	return 0
}
