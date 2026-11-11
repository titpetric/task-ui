package phpfront

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
)

// ModulesDir is the directory of the application tree holding the front end's
// own modules, one per feature of the interface.
const ModulesDir = "modules"

// Modules runs the modinfo.php of every feature of the tree and returns what
// each registered, by name.
//
// None of these reaches the module table a plugin registers into: a feature is
// drawn by the theme's router and not served under Prefix. Its modinfo.php is
// the tree stating its own shape, which a test reads against the directories
// and the schema each one bundles.
func (f *Front) Modules(ctx context.Context) ([]*Module, error) {
	if f == nil {
		return nil, nil
	}
	found, err := Discover(filepath.Join(f.opts.Root, ModulesDir))
	if err != nil {
		return nil, err
	}
	out := make([]*Module, 0, len(found))
	for _, entry := range found {
		if !entry.Startup {
			return nil, fmt.Errorf("%s: %w", entry.File, errNoStartup)
		}
		module, err := startModule(ctx, ModuleOptions{Version: f.opts.Version}, entry)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", entry.File, err)
		}
		out = append(out, module)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out, nil
}

// errNoStartup is returned for a modinfo.php the scanner would never run.
var errNoStartup = errors.New("no @startup annotation, so the module is not registered")
