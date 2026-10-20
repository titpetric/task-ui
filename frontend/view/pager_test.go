package view_test

import (
	"net/url"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
)

func TestNewPager(t *testing.T) {
	t.Parallel()

	query := url.Values{"folder": {"infra"}, "n": {"50"}, "before": {"old"}}
	assert.Equal(t, &view.Pager{Prev: "/runs?after=a&folder=infra&n=50", Next: "/runs?before=b&folder=infra&n=50"}, view.NewPager("/runs", query, "a", "b"))
	assert.Equal(t, &view.Pager{Next: "/runs?before=b&folder=infra&n=50"}, view.NewPager("/runs", query, "", "b"))
	assert.Nil(t, view.NewPager("/runs", query, "", ""))
}
