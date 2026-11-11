package phpscript

import (
	"sort"
	"sync"
	"time"
)

// Public is the directory of a module served as bytes, and Index is what
// answers every path it does not hold.
const (
	Public = "public"
	Index  = "index.php"
)

// Manager holds the modules of the instance: what Discover found, what each
// modinfo.php registered, and a pool of runtimes per module to serve requests
// from. One Manager serves the process.
//
// A nil Manager answers as an instance with no modules, so a handler needs no
// check of its own.
type Manager struct {
	opts Options

	mu      sync.RWMutex
	modules map[string]*Module
	pools   map[string]*pool
	order   []string
}

// New returns the Manager of opts. Nothing is read yet: Start walks the roots.
func New(opts Options) *Manager {
	if opts.Timeout <= 0 {
		opts.Timeout = DefaultTimeout
	}
	if opts.PoolSize <= 0 {
		opts.PoolSize = 2
	}
	return &Manager{
		opts:    opts,
		modules: map[string]*Module{},
		pools:   map[string]*pool{},
	}
}

// Modules returns the registered modules, by name.
func (m *Manager) Modules() []Module {
	if m == nil {
		return nil
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]Module, 0, len(m.order))
	for _, name := range m.order {
		out = append(out, *m.modules[name])
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

// Module returns the module of a name, and whether the instance has it.
func (m *Manager) Module(name string) (*Module, bool) {
	if m == nil {
		return nil, false
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	found, ok := m.modules[name]
	return found, ok
}

// Timeout is how long one render may run, which a handler needs to bound the
// request it passes in.
func (m *Manager) Timeout() time.Duration {
	if m == nil {
		return DefaultTimeout
	}
	return m.opts.Timeout
}

// put records a module and builds its pool, replacing what a name held before:
// a reload reads the modinfo.php again and the runtimes of the old one go with
// the old entry.
func (m *Manager) put(module *Module) {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := module.Name()
	if _, had := m.modules[name]; !had {
		m.order = append(m.order, name)
	}
	m.modules[name] = module
	m.pools[name] = newPool(m.opts, module)
}

// poolOf returns the pool serving a module.
func (m *Manager) poolOf(name string) (*pool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.pools[name]
	return p, ok
}

// reset drops every module, which a reload does before it reads the roots
// again.
func (m *Manager) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.modules = map[string]*Module{}
	m.pools = map[string]*pool{}
	m.order = nil
}
