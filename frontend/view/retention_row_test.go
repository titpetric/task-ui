package view_test

import (
	"strings"
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/retention"
)

// A row says what it holds, lists the ages with the rule's selected,
// and the limits with no limit first and a limit the list lacks at its
// place, selected.
func TestRetentionRow(t *testing.T) {
	t.Parallel()
	r := view.RetentionRow{Area: retention.AreaOf(retention.Recordings), Usage: retention.Usage{Count: 10, Bytes: 10 << 20}, Rule: retention.Rule{Age: retention.Week, Limit: 100 << 20}}
	assert.Equal(t, "10 recordings, 10.0 MB", r.Holds())
	ages := r.Ages()
	assert.Len(t, ages, 6)
	assert.Equal(t, "7 days", ages[1].Label)
	assert.True(t, ages[1].Selected)
	assert.False(t, ages[0].Selected)
	assert.Equal(t, "365 days", ages[4].Label)
	limits := r.Limits()
	assert.Len(t, limits, 7)
	assert.Equal(t, view.Choice{Value: "0", Label: "No limit"}, limits[0])
	assert.Equal(t, view.Choice{Value: "104857600", Label: "100 MB", Selected: true}, limits[1])

	odd := view.RetentionRow{Area: retention.AreaOf(retention.Prompts), Usage: retention.Usage{Count: 3, Bytes: 99}, Rule: retention.Rule{Age: retention.Month, Limit: 2 << 30}}
	assert.Equal(t, "3 threads", odd.Holds())
	limits = odd.Limits()
	assert.Len(t, limits, 8)
	assert.Equal(t, view.Choice{Value: "2147483648", Label: "2 GB", Selected: true}, limits[4])
	assert.False(t, limits[0].Selected)

	none := view.RetentionRow{Area: retention.AreaOf(retention.Signins)}
	assert.Equal(t, "0 rows", none.Holds())
	assert.True(t, none.Limits()[0].Selected)
}

// The page renders the table with a row per area in the table's order,
// the fixed row saying its rule across Keep and Size limit, the
// selects on the saved rule, and no key read log.
func TestRetentionRow_Page(t *testing.T) {
	t.Parallel()
	p := admin(t)
	for _, area := range retention.Areas {
		rule := retention.Rule{Age: retention.Month}
		if area.Kind == retention.Artifacts {
			rule = retention.Rule{Age: retention.Quarter, Limit: 7 << 30}
		}
		p.Retention = append(p.Retention, view.RetentionRow{Area: area, Usage: retention.Usage{Count: 2, Bytes: 4096}, Rule: rule})
	}
	body := render(t, "retention", p)
	last := -1
	for _, area := range retention.Areas {
		at := strings.Index(body, `<td class="key">`+area.Title+`</td>`)
		assert.Greater(t, at, last, area.Title)
		last = at
	}
	assert.Equal(t, len(retention.Kinds), strings.Count(body, `.age" aria-label="Keep `), "a Keep select per rule")
	assert.Equal(t, 3, strings.Count(body, `.limit" aria-label="Size limit of `), "a Size limit select per sized rule")
	assert.NotContains(t, body, "Key read log")
	assert.Contains(t, body, `<select name="artifacts.age" aria-label="Keep Artifact history for">`)
	assert.Contains(t, body, `<option value="90d" selected>90 days</option>`)
	assert.Contains(t, body, `<option value="7516192768" selected>7 GB</option>`, "a saved limit the list lacks is an option, selected")
	assert.Contains(t, body, `<td colspan="2" class="dim">Removed every cleanup, no setting.</td>`)
	assert.Contains(t, body, `<select name="tooling.age" aria-label="Keep Tool history for">`)
	assert.NotContains(t, body, `name="tooling.limit"`)
	assert.Contains(t, body, "2 databases, 4.0 KB")
	assert.Contains(t, body, "2 recordings, 4.0 KB")
	assert.Contains(t, body, "2 threads<")
	assert.Contains(t, body, `Applied by <a href="/runs?folder=_&amp;job=@cleanup" title="The runs of @cleanup">@cleanup</a>, which <span class="faint">has not run yet</span>.`, "the lead links the job; no cleanup yet")
	assert.NotContains(t, body, "Last cleanup")
	assert.Contains(t, body, `data-api="/api/retention" data-method="PUT"`)
}
