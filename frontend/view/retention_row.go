package view

import (
	"slices"
	"strconv"

	"github.com/titpetric/task-ui/retention"
)

// RetentionRow is one area of the Data retention table: what it holds
// and the rule that keeps it, which the row's selects set. A fixed area
// shows its text in place of the selects.
type RetentionRow struct {
	Area  retention.Area
	Usage retention.Usage
	Rule  retention.Rule
}

// Holds reads what the area holds, the bytes too for a sized one.
func (r RetentionRow) Holds() string {
	return r.Area.Holds(r.Usage)
}

// Ages are the ages the Keep select lists, the rule's selected.
func (r RetentionRow) Ages() []Choice {
	ages := make([]Choice, 0, len(retention.Ages))
	for _, age := range retention.Ages {
		ages = append(ages, Choice{Value: age, Label: retention.Label(age), Selected: age == r.Rule.Age})
	}
	return ages
}

// Limits are the size limits to pick from, no limit first, the rule's selected, a limit the list lacks at its place.
func (r RetentionRow) Limits() []Choice {
	sizes := retention.Limits
	if r.Rule.Limit > 0 && !slices.Contains(sizes, r.Rule.Limit) {
		sizes = append(slices.Clone(sizes), r.Rule.Limit)
		slices.Sort(sizes)
	}
	limits := []Choice{{Value: "0", Label: "No limit", Selected: r.Rule.Limit == 0}}
	for _, size := range sizes {
		limits = append(limits, Choice{Value: strconv.FormatInt(size, 10), Label: retention.SizeLabel(size), Selected: size == r.Rule.Limit})
	}
	return limits
}
