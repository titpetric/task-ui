package view

import (
	"github.com/titpetric/task-ui/retention"
)

// RetentionBlock is one kind of data on the Data retention page: what it
// holds and the rule that keeps it.
type RetentionBlock struct {
	Kind  string
	Title string
	Desc  string
	// One and Many name an item of the kind: run, runs.
	One, Many string
	Usage     retention.Usage
	Rule      retention.Rule
}

// Average is the average size of an item the kind holds, which the page
// estimates a size limit by; zero without any.
func (b RetentionBlock) Average() int64 {
	if b.Usage.Count == 0 {
		return 0
	}
	return b.Usage.Bytes / int64(b.Usage.Count)
}

// Estimate says about how many items the rule's size limit holds: "about
// 120 runs", empty without a limit or an average.
func (b RetentionBlock) Estimate() string {
	n := retention.Estimate(b.Rule.Limit, b.Usage.Count, b.Usage.Bytes)
	if n == 0 {
		return ""
	}
	return "about " + retention.Usage{Count: n}.Name(b.One, b.Many)
}

// LimitText is the size limit as its field shows it, empty for none.
func (b RetentionBlock) LimitText() string {
	if b.Rule.Limit == 0 {
		return ""
	}
	return retention.FormatSize(b.Rule.Limit)
}

// Ages are the ages the page lists, each with its label, the rule's checked.
func (b RetentionBlock) Ages() []struct {
	Value, Label string
	Checked      bool
} {
	labels := map[string]string{retention.Day: "24 hours", retention.Week: "7 days", retention.Month: "31 days", retention.Never: "Never"}
	ages := []struct {
		Value, Label string
		Checked      bool
	}{}
	for _, age := range retention.Ages {
		ages = append(ages, struct {
			Value, Label string
			Checked      bool
		}{age, labels[age], age == b.Rule.Age})
	}
	return ages
}

// Sizes are the size limits the page lists beside its field.
func (b RetentionBlock) Sizes() []string {
	return []string{"100 MB", "1 GB", "10 GB"}
}
