package askhuman_test

import (
	"testing"

	"github.com/titpetric/task-ui/feature/askhuman"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSubmission(t *testing.T) {
	t.Parallel()
	f := &askhuman.Form{Fields: []askhuman.Field{
		{Key: "hosts", Component: "multi", Source: "hosts"},
		{Key: "count", Component: "number"},
	}}
	s, err := f.Bind(map[string]string{"hosts": "a,b", "count": "x"})
	assert.ErrorIs(t, err, askhuman.ErrInvalid)
	assert.Equal(t, &askhuman.Submission{
		Values: map[string]string{"hosts": "a, b"},
		List:   map[string][]string{"hosts": {"a", "b"}},
		Errors: map[string]string{"count": `"x" is no number`},
	}, s)
}
