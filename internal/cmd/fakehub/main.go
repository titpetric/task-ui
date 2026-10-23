// Command fakehub is a Beszel hub for the inventory's screenshots: three
// hosts, one with an extra disk and one down, their latest minute, their
// containers, and a month of eight hour records drawn from smooth curves,
// a gap of two days in one host's and the down host's ending when it
// went down.
//
//	go run ./internal/cmd/fakehub -listen 127.0.0.1:3190
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:3190", "the address to serve on")
	flag.Parse()
	if err := http.ListenAndServe(*listen, newHub(time.Now)); err != nil {
		log.Fatal(err)
	}
}

// host is a host of the hub: its system row, how busy it runs and what
// it holds.
type host struct {
	ID, Name, Addr string
	Up             bool
	Threads        int
	// CPU and Mem are the mean CPU and memory in use, in percent; Memory
	// and Disk the RAM and the root disk in GB, DiskUsed the root's use.
	CPU, Mem       float64
	Memory, Disk   float64
	DiskUsed       float64
	Extra          map[string][2]float64
	Containers     []container
	DownFor, GapAt time.Duration
}

// container is a container of a host: its name, image and status, CPU in
// percent of the host and memory in MB.
type container struct {
	Name, Image, Status string
	CPU, Mem            float64
}

var hosts = []host{
	{
		ID: "pve0000000000a", Name: "pve", Addr: "10.0.0.2", Up: true, Threads: 16, CPU: 31, Mem: 58,
		Memory: 62.7, Disk: 93.9, DiskUsed: 41.2, Extra: map[string][2]float64{"data": {1862.6, 1104.3}},
		Containers: []container{
			{"postgres", "postgres:17", "Up 12 days", 6.4, 2148},
			{"prometheus", "prom/prometheus:v3.1.0", "Up 12 days", 4.1, 1830},
			{"grafana", "grafana/grafana:11.4.0", "Up 12 days", 1.2, 412},
			{"minio", "minio/minio:latest", "Up 3 days", 2.7, 980},
			{"loki", "grafana/loki:3.3.2", "Up 12 days", 1.9, 760},
			{"caddy", "caddy:2.9", "Up 12 days", 0.3, 64},
		},
	},
	{
		ID: "dev10000000000", Name: "dev1", Addr: "10.0.0.11", Up: true, Threads: 4, CPU: 22, Mem: 64,
		Memory: 7.7, Disk: 195.8, DiskUsed: 88.4, GapAt: 11 * 24 * time.Hour,
		Containers: []container{
			{"task-ui", "registry.cdn.si/task-ui:latest", "Up 2 hours", 3.8, 286},
			{"forgejo", "codeberg.org/forgejo/forgejo:9", "Up 5 days", 1.1, 498},
			{"redis", "redis:7-alpine", "Up 5 days", 0.4, 38},
			{"caddy", "caddy:2.9", "Up 5 days", 0.2, 41},
		},
	},
	{
		ID: "docker1000000a", Name: "docker1", Addr: "10.0.0.12", Threads: 4, CPU: 12, Mem: 41,
		Memory: 7.7, Disk: 97.9, DiskUsed: 30.5, DownFor: 4 * 24 * time.Hour,
		Containers: []container{{"nginx", "nginx:1.27", "Up 9 days", 0.5, 22}},
	},
}

// newHub serves the collections the inventory reads, at the time now
// returns.
func newHub(now func() time.Time) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/collections/systems/records", func(w http.ResponseWriter, _ *http.Request) {
		var items []any
		for _, h := range hosts {
			status, cpu, mem := "down", 0.0, 0.0
			if h.Up {
				status, cpu, mem = "up", h.CPU, h.Mem
			}
			items = append(items, map[string]any{
				"id": h.ID, "name": h.Name, "host": h.Addr, "status": status,
				"info": map[string]any{"t": h.Threads, "u": 14 * 24 * 3600, "cpu": cpu, "mp": mem, "dp": h.DiskUsed * 100 / h.Disk, "v": "0.18.4"},
			})
		}
		write(w, items)
	})
	mux.HandleFunc("/api/collections/system_stats/records", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		var items []any
		for _, h := range hosts {
			if !matches(filter, h) {
				continue
			}
			if strings.Contains(filter, "type='1m'") {
				if h.Up {
					items = append(items, map[string]any{"system": h.ID, "created": stamp(now()), "stats": h.stats(0)})
				}
				continue
			}
			for _, at := range h.month(now(), filter) {
				items = append(items, map[string]any{"system": h.ID, "created": stamp(at), "stats": h.stats(now().Sub(at))})
			}
		}
		write(w, items)
	})
	mux.HandleFunc("/api/collections/containers/records", func(w http.ResponseWriter, _ *http.Request) {
		var items []any
		for _, h := range hosts {
			for i, c := range h.Containers {
				if h.Up {
					items = append(items, map[string]any{"system": h.ID, "name": c.Name, "image": c.Image, "status": c.Status, "cpu": c.CPU, "memory": c.Mem, "net": 18000 * float64(len(h.Containers)-i)})
				}
			}
		}
		write(w, items)
	})
	mux.HandleFunc("/api/collections/container_stats/records", func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		var items []any
		for _, h := range hosts {
			if !matches(filter, h) {
				continue
			}
			for _, at := range h.month(now(), filter) {
				age := now().Sub(at)
				var stats []any
				for i, c := range h.Containers {
					k := wave(float64(i)+1, age)
					stats = append(stats, map[string]any{"n": c.Name, "c": c.CPU * (0.6 + 0.6*k), "m": c.Mem * (0.75 + 0.35*wave(float64(i)+3.5, age))})
				}
				items = append(items, map[string]any{"system": h.ID, "created": stamp(at), "stats": stats})
			}
		}
		write(w, items)
	})
	return mux
}

// matches reports whether a filter takes the host's records: one that
// names no system takes every host's.
func matches(filter string, h host) bool {
	return !strings.Contains(filter, "system=") || strings.Contains(filter, "system='"+h.ID+"'")
}

// month lists the times of the host's records the filter selects, oldest
// first: every eight hours of the last thirty days for the 480m records,
// none of another type, none in a gap or after the host went down.
func (h host) month(now time.Time, filter string) []time.Time {
	if !strings.Contains(filter, "type='480m'") {
		return nil
	}
	step := 8 * time.Hour
	end := now.Truncate(step)
	var times []time.Time
	for at := end.Add(-30 * 24 * time.Hour).Add(step); !at.After(end); at = at.Add(step) {
		age := now.Sub(at)
		if age < h.DownFor || (h.GapAt > 0 && age < h.GapAt && age > h.GapAt-2*24*time.Hour) {
			continue
		}
		times = append(times, at)
	}
	return times
}

// stats is the host's record of an age: its means swayed by a daily and a
// weekly curve, the disks filling slowly over the month.
func (h host) stats(age time.Duration) map[string]any {
	k := wave(float64(h.Threads), age)
	cpu := h.CPU * (0.55 + 0.9*k)
	memUsed := h.Memory * h.Mem / 100 * (0.85 + 0.3*wave(2.5, age))
	grow := 1 - age.Hours()/(24*30)*0.08
	stats := map[string]any{
		"cpu": round(cpu), "m": h.Memory, "mu": round(memUsed),
		"d": h.Disk, "du": round(h.DiskUsed * grow),
		"b": []float64{round(2.1e5 * (0.4 + k)), round(6.4e5 * (0.3 + wave(7, age)))},
	}
	if len(h.Extra) > 0 {
		efs := map[string]any{}
		for name, size := range h.Extra {
			efs[name] = map[string]float64{"d": size[0], "du": round(size[1] * grow)}
		}
		stats["efs"] = efs
	}
	return stats
}

// wave is a value from 0 to 1 that sways with a record's age, daily and
// weekly, its phase set by seed.
func wave(seed float64, age time.Duration) float64 {
	days := age.Hours() / 24
	v := 0.5 + 0.25*math.Sin(days*2*math.Pi/7+seed) + 0.2*math.Sin(days*2*math.Pi/2.3+seed*1.7) + 0.05*math.Sin(days*9+seed)
	return math.Max(0, math.Min(1, v))
}

// round keeps two decimals, as the hub does.
func round(v float64) float64 {
	return math.Round(v*100) / 100
}

// stamp is a time as PocketBase writes it.
func stamp(at time.Time) string {
	return at.UTC().Format("2006-01-02 15:04:05.000Z")
}

// write answers a page of records, every record on the first.
func write(w http.ResponseWriter, items []any) {
	if items == nil {
		items = []any{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]any{"page": 1, "items": items}); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
