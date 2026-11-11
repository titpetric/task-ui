package phpscript

import (
	"context"
)

// ModuleFrom returns the module of the context, and whether one is there. A
// context without one is a binding the Manager did not set up, which it refuses.
func ModuleFrom(ctx context.Context) (*Module, bool) {
	m, ok := ctx.Value(moduleKey{}).(*Module)
	return m, ok && m != nil
}
