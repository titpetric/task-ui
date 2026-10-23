package retention

import (
	"strconv"
)

// The kinds of data the rules keep, in the order the page shows them.
const (
	Recordings  = "recordings"
	Artifacts   = "artifacts"
	SecretReads = "secret_reads"
	KeyReads    = "key_reads"
)

// Kinds are the kinds of data, in the order the page shows them.
var Kinds = []string{Recordings, Artifacts, SecretReads, KeyReads}

// Settings are the rules of every kind of data, as the Data retention
// page sets them.
type Settings struct {
	Recordings  Rule `json:"recordings"`
	Artifacts   Rule `json:"artifacts"`
	SecretReads Rule `json:"secret_reads"`
	KeyReads    Rule `json:"key_reads"`
}

// Defaults keep a recording 31 days and the rest for ever, with no size limit.
func Defaults() Settings {
	return Settings{Recordings: Rule{Age: Month}, Artifacts: Rule{Age: Never}, SecretReads: Rule{Age: Never}, KeyReads: Rule{Age: Never}}
}

// Load reads the settings from stored values, the defaults for what is
// missing or wrong.
func Load(values map[string]string) Settings {
	s := Defaults()
	for _, kind := range Kinds {
		rule := s.Rule(kind)
		if age := values["retention."+kind+".age"]; age != "" {
			rule.Age = age
		}
		if limit, err := strconv.ParseInt(values["retention."+kind+".limit"], 10, 64); err == nil {
			rule.Limit = limit
		}
		if rule.Valid() {
			s.set(kind, rule)
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
	return s.Recordings.Valid() && s.Artifacts.Valid() && s.SecretReads.Valid() && s.KeyReads.Valid()
}

// Rule returns the rule of a kind.
func (s Settings) Rule(kind string) Rule {
	switch kind {
	case Recordings:
		return s.Recordings
	case Artifacts:
		return s.Artifacts
	case KeyReads:
		return s.KeyReads
	}
	return s.SecretReads
}

func (s *Settings) set(kind string, rule Rule) {
	switch kind {
	case Recordings:
		s.Recordings = rule
	case Artifacts:
		s.Artifacts = rule
	case KeyReads:
		s.KeyReads = rule
	default:
		s.SecretReads = rule
	}
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
