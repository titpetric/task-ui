// Package killswitch stops a test binary before the kernel does.
//
// GOMEMLIMIT is a target and not a bound: it tells the collector to work
// harder, and a binary that allocates faster than the collector frees
// grows past it anyway. What follows is the OOM killer, which names no
// test, writes no stack and leaves a run that reports "signal: killed"
// against a package of eighty tests.
//
// This watches the process's own resident memory, reports it, and panics
// the first reading over the limit. A panic is the point: it runs with
// every goroutine's stack printed, so the reading that kills the process
// comes with the test that was running and the call tree underneath it.
package killswitch

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

// Limit is the ceiling a test binary gets when nothing says otherwise.
// The machines this runs on have less than twice that, so a binary over
// it is already taking the suite down with it.
const Limit uint64 = 8 << 30

// Names of what the environment can say. KillEnv takes a byte size the
// way GOMEMLIMIT does -- 8GiB, 512MiB, 2000000 -- or "off". TraceEnv
// turns on the reading printed every interval; without it only the peak
// and the reading that kills are printed.
const (
	KillEnv    = "TASK_UI_TEST_MEM_KILL"
	TraceEnv   = "TASK_UI_TEST_MEM_TRACE"
	ProfileEnv = "TASK_UI_TEST_MEM_PROFILE"
)

// Options is what a watcher does. The zero value watches the default
// limit every second and writes to stdout.
type Options struct {
	// Limit is the resident size that panics the process, Limit by
	// default. A negative limit watches and reports without killing.
	Limit int64
	// Every is how often memory is read, 100ms by default. A test binary
	// that leaks allocates half a gigabyte a second, measured, so a
	// watcher reading once a second overshoots its limit by that much
	// and reports a peak that is as far under the truth.
	Every time.Duration
	// Report is how often a reading is printed under Trace, a second by
	// default: the readings are for a person to watch, and ten a second
	// is not.
	Report time.Duration
	// Trace prints the readings rather than only the peak and the last.
	Trace bool
	// Profile is a file a heap profile is written to before the panic.
	// Empty writes none.
	Profile string
	// Out is where the readings go, os.Stdout by default.
	Out io.Writer
	// Kill is what the reading over the limit does, a panic by default.
	// A test of this package passes its own.
	Kill func(reading string)
}

// Watch starts the watcher the environment asks for and answers the
// function that stops it, which prints the peak. It is what a TestMain
// calls: no argument, nothing to keep in sync with the suite's own
// limits, and TASK_UI_TEST_MEM_KILL=off for a run that wants none.
func Watch() func() {
	limit, ok := limitFromEnv()
	if !ok {
		return func() {}
	}
	_, trace := os.LookupEnv(TraceEnv)
	return Start(Options{
		Limit:   limit,
		Trace:   trace,
		Profile: os.Getenv(ProfileEnv),
	})
}

// Start watches as the options say and answers the function that stops
// it. Calling the stopper twice is the same as calling it once.
func Start(o Options) func() {
	if o.Limit == 0 {
		o.Limit = int64(Limit)
	}
	if o.Every <= 0 {
		o.Every = 100 * time.Millisecond
	}
	if o.Report <= 0 {
		o.Report = time.Second
	}
	if o.Out == nil {
		o.Out = os.Stdout
	}
	if o.Kill == nil {
		o.Kill = kill(o.Profile)
	}

	w := &watcher{o: o, done: make(chan struct{}), stopped: make(chan struct{})}
	go w.run()
	return w.stop
}

// watcher is one running watch.
type watcher struct {
	o       Options
	peak    atomic.Uint64
	said    time.Time
	done    chan struct{}
	stopped chan struct{}
	once    atomic.Bool
}

// run reads memory every interval until it is stopped, and does not
// return once the limit is passed: Kill panics, and a panic that unwound
// this goroutine quietly would leave the process alive and over.
func (w *watcher) run() {
	defer close(w.stopped)
	tick := time.NewTicker(w.o.Every)
	defer tick.Stop()
	for {
		w.read()
		select {
		case <-w.done:
			return
		// lifetime: one reading per interval until the stopper is called
		// or the limit is passed.
		case <-tick.C:
		}
	}
}

// read takes one reading, records the peak, prints it if the report
// interval has passed, and kills the process if it is over.
func (w *watcher) read() {
	r := reading()
	if r.rss > w.peak.Load() {
		w.peak.Store(r.rss)
	}
	over := w.o.Limit > 0 && r.rss > uint64(w.o.Limit)
	if w.o.Trace && !over && time.Since(w.said) >= w.o.Report {
		w.said = time.Now()
		fmt.Fprintf(w.o.Out, "killswitch: %s\n", r)
	}
	if over {
		line := fmt.Sprintf("%s over the %s limit", r, size(uint64(w.o.Limit)))
		fmt.Fprintf(w.o.Out, "killswitch: %s\n", line)
		w.o.Kill(line)
	}
}

// stop ends the watch and prints the peak, which is the number a run is
// read for: what one package of tests cost at its worst.
func (w *watcher) stop() {
	if w.once.Swap(true) {
		return
	}
	close(w.done)
	<-w.stopped
	fmt.Fprintf(w.o.Out, "killswitch: peak rss %s\n", size(w.peak.Load()))
}

// kill writes a heap profile where it was asked for, turns the traceback
// up to every goroutine, and panics. The stacks are the reason this is a
// panic and not an exit: they name the test that was running.
func kill(profile string) func(string) {
	return func(line string) {
		if profile != "" {
			if f, err := os.Create(profile); err == nil {
				runtime.GC()
				if err := pprof.Lookup("heap").WriteTo(f, 0); err != nil {
					fmt.Fprintf(os.Stdout, "killswitch: the heap profile: %v\n", err)
				}
				f.Close()
				fmt.Fprintf(os.Stdout, "killswitch: heap profile in %s\n", profile)
			}
		}
		debug.SetTraceback("all")
		panic("killswitch: " + line)
	}
}

// stat is one reading of what the process holds.
type stat struct {
	rss        uint64
	heap, sys  uint64
	goroutines int
}

// String is the line a reading prints as.
func (s stat) String() string {
	return fmt.Sprintf("rss %s, heap %s, sys %s, %d goroutines",
		size(s.rss), size(s.heap), size(s.sys), s.goroutines)
}

// reading reads the process's resident size and the runtime's own
// numbers. Both are printed because they answer different questions: a
// heap that is flat while the resident size grows is memory the Go
// collector does not own, which is cgo, mmap or a leaked file.
func reading() stat {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s := stat{heap: m.HeapAlloc, sys: m.Sys, goroutines: runtime.NumGoroutine()}
	s.rss = rss()
	if s.rss == 0 {
		// No /proc, so the runtime's own total is the best there is. It
		// counts what Go reserved rather than what is resident, which is
		// the larger of the two and so fails safe.
		s.rss = m.Sys
	}
	return s
}

// rss is the resident size out of /proc/self/statm, whose second field
// is the resident page count. It answers zero where there is no procfs.
func rss() uint64 {
	b, err := os.ReadFile("/proc/self/statm")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(b))
	if len(fields) < 2 {
		return 0
	}
	pages, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return 0
	}
	return pages * uint64(os.Getpagesize())
}

// limitFromEnv reads the limit the environment asks for. It answers
// false for "off", and the default for anything it cannot read, because
// a typo in a limit should not quietly remove the limit.
func limitFromEnv() (int64, bool) {
	raw, ok := os.LookupEnv(KillEnv)
	if !ok || raw == "" {
		return int64(Limit), true
	}
	if strings.EqualFold(raw, "off") || raw == "0" {
		return 0, false
	}
	n, err := ParseSize(raw)
	if err != nil {
		fmt.Fprintf(os.Stdout, "killswitch: %s=%s: %v, using %s\n", KillEnv, raw, err, size(Limit))
		return int64(Limit), true
	}
	return n, true
}

// ParseSize reads a byte size the way GOMEMLIMIT is written: a number,
// optionally followed by B, KiB, MiB, GiB or TiB.
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	digits := 0
	for digits < len(s) && s[digits] >= '0' && s[digits] <= '9' {
		digits++
	}
	if digits == 0 {
		return 0, fmt.Errorf("no number in %q", s)
	}
	n, err := strconv.ParseInt(s[:digits], 10, 64)
	if err != nil {
		return 0, err
	}
	unit := uint64(1)
	switch strings.ToLower(strings.TrimSpace(s[digits:])) {
	case "", "b":
	case "kib":
		unit = 1 << 10
	case "mib":
		unit = 1 << 20
	case "gib":
		unit = 1 << 30
	case "tib":
		unit = 1 << 40
	default:
		return 0, fmt.Errorf("no unit %q", s[digits:])
	}
	if n > 0 && uint64(n) > uint64(1<<62)/unit {
		return 0, fmt.Errorf("%q does not fit", s)
	}
	return n * int64(unit), nil
}

// size writes a byte count the way the limit is written, so a reading
// and the limit it is compared against read the same.
func size(n uint64) string {
	switch {
	case n >= 1<<30:
		return fmt.Sprintf("%.2fGiB", float64(n)/(1<<30))
	case n >= 1<<20:
		return fmt.Sprintf("%.0fMiB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%.0fKiB", float64(n)/(1<<10))
	default:
		return fmt.Sprintf("%dB", n)
	}
}
