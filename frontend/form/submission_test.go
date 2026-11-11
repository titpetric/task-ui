package form_test

import (
	"testing"

	"github.com/titpetric/task-ui/form"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestSubmission(t *testing.T) {
	t.Parallel()
	f := &form.Form{Fields: []form.Field{
		{Key: "hosts", Component: "multi", Source: "hosts"},
		{Key: "count", Component: "number"},
	}}
	s, err := f.Bind(map[string]string{"hosts": "a,b", "count": "x"})
	assert.ErrorIs(t, err, form.ErrInvalid)
	assert.Equal(t, &form.Submission{
		Values: map[string]string{"hosts": "a, b"},
		List:   map[string][]string{"hosts": {"a", "b"}},
		Errors: map[string]string{"count": `"x" is no number`},
	}, s)
}
