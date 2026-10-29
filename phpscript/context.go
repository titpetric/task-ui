package phpscript

import (
	"context"
)

// moduleKey is the context key the module travels under. A type of its own, so
// nothing else in the process can set or read it by accident.
type moduleKey struct{}

// WithModule returns a context naming the module whose runtime is about to
// run. A binding reads the module there and takes no argument for it.
func WithModule(ctx context.Context, m *Module) context.Context {
	return context.WithValue(ctx, moduleKey{}, m)
}
