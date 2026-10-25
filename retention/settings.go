package retention

import (
	"strconv"
)

// The kinds of data a rule keeps, in the order the page shows them and
// Clean applies them.
const (
	Recordings   = "recordings"
	Runs         = "runs"
	Prompts      = "prompts"
	WorkflowRuns = "workflow_runs"
	Artifacts    = "artifacts"
	SecretReads  = "secret_reads"
	// Tooling is the tool history atkins keeps in each workspace's home
	// and the instance's own, which the cleanup prunes through atkins.
	Tooling = "tooling"
)

// Kinds are the kinds of data with a rule, in the order the page shows them.
var Kinds = []string{Recordings, Runs, Prompts, WorkflowRuns, Artifacts, SecretReads, Tooling}

// Settings are the rules of every kind of data, by kind, as the Data
// retention page sets them. A kind the map lacks keeps the default.
type Settings map[string]Rule

// Defaults keep every kind 31 days, with no size limit.
func Defaults() Settings {
	s := Settings{}
	for _, kind := range Kinds {
		s[kind] = Rule{Age: Month}
	}
	return s
}

// Load reads the settings from stored values, with the defaults for what is missing or wrong.
func Load(values map[string]string) Settings {
	s := Defaults()
	for _, kind := range Kinds {
		rule := s[kind]
		if age := values["retention."+kind+".age"]; age != "" {
			rule.Age = age
		}
		if limit, err := strconv.ParseInt(values["retention."+kind+".limit"], 10, 64); err == nil {
			rule.Limit = limit
		}
		if rule.Valid() {
			s[kind] = rule
		}
	}
	return s
}

// Values are the settings as values to store.
func (s Settings) Values() map[string]string {
	values := map[string]string{}
	for _, kind := range Kinds {
		rule := s.Rule(kind)
		values["retention."+kind+".age"] = rule.Age
		values["retention."+kind+".limit"] = strconv.FormatInt(rule.Limit, 10)
	}
	return values
}

// Valid reports whether every rule is.
func (s Settings) Valid() bool {
	for _, kind := range Kinds {
		if !s.Rule(kind).Valid() {
			return false
		}
	}
	return true
}

// Rule returns the rule of a kind, the default for one the settings lack.
func (s Settings) Rule(kind string) Rule {
	if rule, ok := s[kind]; ok {
		return rule
	}
	return Rule{Age: Month}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
