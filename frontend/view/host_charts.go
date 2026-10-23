package view

import (
	"math"
	"sort"
	"time"

	"github.com/titpetric/task-ui/beszel"
	"github.com/titpetric/task-ui/hostinfo"
)

// ChartTop is how many containers a container chart draws: the busiest
// of the month.
const ChartTop = 5

// HostCharts draws a host's month to now: CPU, memory against its RAM, each disk's fill and the traffic.
func HostCharts(system beszel.System, samples []beszel.Sample, now time.Time) []Chart {
	from := now.Add(-beszel.Month)
	cpu := ChartSeries{Name: "CPU"}
	memory := ChartSeries{Name: "Used"}
	sent := ChartSeries{Name: "Sent", Color: 0}
	recv := ChartSeries{Name: "Received", Color: 1}
	disks := map[string]*ChartSeries{"/": {Name: "/"}}
	ram := 0.0
	for _, s := range samples {
		cpu.Points = append(cpu.Points, ChartPoint{At: s.At, V: s.Stats.CPU})
		memory.Points = append(memory.Points, ChartPoint{At: s.At, V: s.Stats.MemoryUsed})
		ram = math.Max(ram, s.Stats.Memory)
		out, in := s.Stats.Net()
		sent.Points = append(sent.Points, ChartPoint{At: s.At, V: out})
		recv.Points = append(recv.Points, ChartPoint{At: s.At, V: in})
		disks["/"].Points = append(disks["/"].Points, ChartPoint{At: s.At, V: percent(s.Stats.DiskUsed, s.Stats.Disk)})
		for name, fs := range s.Stats.ExtraFs {
			if disks[name] == nil {
				disks[name] = &ChartSeries{Name: name}
			}
			disks[name].Points = append(disks[name].Points, ChartPoint{At: s.At, V: percent(fs.Used, fs.Total)})
		}
	}
	names := make([]string, 0, len(disks))
	for name := range disks {
		names = append(names, name)
	}
	sort.Strings(names)
	var disk []ChartSeries
	for i, name := range names {
		disks[name].Color = i % hostinfo.Colors
		disk = append(disk, *disks[name])
	}
	return []Chart{
		NewChart("CPU", ChartPercent, from, now, cpu),
		NewChart("Memory", ChartUnit{Max: ram, Scale: ChartGiB.Scale}, from, now, memory),
		NewChart("Disk", ChartPercent, from, now, disk...),
		NewChart("Network", ChartRate, from, now, sent, recv),
	}
}

// ContainerCharts draws a month of the ChartTop containers of most mean memory, and of most CPU.
func ContainerCharts(now time.Time, samples []beszel.ContainerSample) []Chart {
	from := now.Add(-beszel.Month)
	memory := map[string][]ChartPoint{}
	cpu := map[string][]ChartPoint{}
	for _, s := range samples {
		for _, c := range s.Containers {
			memory[c.Name] = append(memory[c.Name], ChartPoint{At: s.At, V: c.Memory})
			cpu[c.Name] = append(cpu[c.Name], ChartPoint{At: s.At, V: c.CPU})
		}
	}
	return []Chart{
		NewChart("Containers by memory", ChartMB, from, now, busiest(memory)...),
		NewChart("Containers by CPU", ChartUnit{Scale: ChartPercent.Scale}, from, now, busiest(cpu)...),
	}
}

// ClusterCharts draws the month's busy cores and memory of the pools' hosts, stacked, in their pools' colours.
func ClusterCharts(cluster *hostinfo.Cluster, systems []beszel.System, history map[string][]beszel.Sample, now time.Time) []Chart {
	from := now.Add(-beszel.Month)
	step := beszelStep(history)
	byName := map[string]beszel.System{}
	for _, system := range systems {
		byName[system.Name] = system
	}
	var times []time.Time
	seen := map[time.Time]bool{}
	for _, host := range cluster.Hosts {
		for _, s := range history[byName[host.Name].ID] {
			if at := s.At.Truncate(step); !seen[at] {
				seen[at] = true
				times = append(times, at)
			}
		}
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })
	var cores, memory []ChartSeries
	for _, host := range cluster.Hosts {
		if host.Left {
			continue
		}
		system := byName[host.Name]
		busy, used := map[time.Time]float64{}, map[time.Time]float64{}
		for _, s := range history[system.ID] {
			at := s.At.Truncate(step)
			busy[at] = s.Stats.CPU / 100 * float64(system.CPUs())
			used[at] = s.Stats.MemoryUsed
		}
		c, m := ChartSeries{Name: host.Name, Color: host.Color}, ChartSeries{Name: host.Name, Color: host.Color}
		for _, at := range times {
			c.Points = append(c.Points, ChartPoint{At: at, V: valueAt(at, busy)})
			m.Points = append(m.Points, ChartPoint{At: at, V: valueAt(at, used)})
		}
		cores, memory = append(cores, c), append(memory, m)
	}
	return []Chart{
		NewStackedChart("Busy cores", ChartCores, from, now, cores...),
		NewStackedChart("Memory used", ChartGiB, from, now, memory...),
	}
}

// beszelStep is the span of the records of a history: the shortest
// between two of a host, the hub's eight hours without two.
func beszelStep(history map[string][]beszel.Sample) time.Duration {
	step := time.Duration(0)
	for _, samples := range history {
		for i := 1; i < len(samples); i++ {
			if d := samples[i].At.Sub(samples[i-1].At); d >= time.Hour && (step == 0 || d < step) {
				step = d
			}
		}
	}
	if step == 0 {
		return chartStep
	}
	return step.Round(time.Hour)
}

// valueAt is a host's value at a time, NaN where it has none.
func valueAt(at time.Time, values map[time.Time]float64) float64 {
	if v, ok := values[at]; ok {
		return v
	}
	return math.NaN()
}

// busiest returns the ChartTop series of the highest mean, highest
// first, each in its colour.
func busiest(points map[string][]ChartPoint) []ChartSeries {
	type mean struct {
		name string
		avg  float64
	}
	var means []mean
	for name, ps := range points {
		sum := 0.0
		for _, p := range ps {
			sum += p.V
		}
		means = append(means, mean{name, sum / float64(len(ps))})
	}
	sort.Slice(means, func(i, j int) bool {
		if means[i].avg != means[j].avg {
			return means[i].avg > means[j].avg
		}
		return means[i].name < means[j].name
	})
	var series []ChartSeries
	for i, m := range means {
		if i == ChartTop {
			break
		}
		series = append(series, ChartSeries{Name: m.name, Color: i, Points: points[m.name]})
	}
	return series
}

// percent is the share of total in use, 0 of no total.
func percent(used, total float64) float64 {
	if total <= 0 {
		return 0
	}
	return used / total * 100
}
