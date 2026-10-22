// Package config holds the configuration task-ui embeds: the rules of
// permissions.yml, which every permission check names, and context.md,
// what task-ui is, for an agent.
package config

import (
	"embed"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed permissions.yml context.md
var files embed.FS

// All is the rule that stands for every rule of a section in a policy.
// No check names it.
const All = "all"

// Permissions are the rules a policy sets and a check names: a path
// <section>/<rule> for each section and rule, and a path <permission>
// for each application permission.
type Permissions struct {
	// Rules every section gets, All among them.
	Rules []Entry
	// Sections of the application, each with every rule.
	Sections []Entry
	// Permissions are application wide, beyond the sections.
	Permissions []Entry
}

var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

// Parse reads and validates a permissions file. Every list needs an
// entry, keys are lower case words, unique across the sections and the
// permissions, and the rules hold all, view, edit and delete.
func Parse(data []byte) (*Permissions, error) {
	var doc struct {
		Rules       yaml.Node `yaml:"rules"`
		Sections    yaml.Node `yaml:"sections"`
		Permissions yaml.Node `yaml:"permissions"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("permissions: %w", err)
	}
	p := &Permissions{}
	var err error
	if p.Rules, err = entries("rules", &doc.Rules); err != nil {
		return nil, err
	}
	if p.Sections, err = entries("sections", &doc.Sections); err != nil {
		return nil, err
	}
	if p.Permissions, err = entries("permissions", &doc.Permissions); err != nil {
		return nil, err
	}
	rules := map[string]bool{}
	for _, rule := range p.Rules {
		rules[rule.Key] = true
	}
	for _, need := range []string{All, "view", "edit", "delete"} {
		if !rules[need] {
			return nil, fmt.Errorf("permissions: rules: no %q", need)
		}
	}
	seen := map[string]bool{}
	for _, entry := range append(append([]Entry{}, p.Sections...), p.Permissions...) {
		if seen[entry.Key] {
			return nil, fmt.Errorf("permissions: %q is both a section and a permission", entry.Key)
		}
		seen[entry.Key] = true
	}
	return p, nil
}

// entries reads a mapping of keys to titles, in order.
func entries(name string, node *yaml.Node) ([]Entry, error) {
	if node.Kind != yaml.MappingNode || len(node.Content) == 0 {
		return nil, fmt.Errorf("permissions: %s: a mapping of keys to titles is needed", name)
	}
	var list []Entry
	seen := map[string]bool{}
	for i := 0; i+1 < len(node.Content); i += 2 {
		key, title := node.Content[i], node.Content[i+1]
		if !keyPattern.MatchString(key.Value) {
			return nil, fmt.Errorf("permissions: %s: line %d: key %q is no lower case word", name, key.Line, key.Value)
		}
		if title.Kind != yaml.ScalarNode || strings.TrimSpace(title.Value) == "" {
			return nil, fmt.Errorf("permissions: %s: line %d: %q needs a title", name, key.Line, key.Value)
		}
		if seen[key.Value] {
			return nil, fmt.Errorf("permissions: %s: line %d: %q twice", name, key.Line, key.Value)
		}
		seen[key.Value] = true
		list = append(list, Entry{Key: key.Value, Title: title.Value})
	}
	return list, nil
}

var embedded = mustParse("permissions.yml")

// mustParse reads an embedded file; a bad one is a build that cannot
// run, which the tests catch.
func mustParse(name string) *Permissions {
	data, err := files.ReadFile(name)
	if err == nil {
		var p *Permissions
		if p, err = Parse(data); err == nil {
			return p
		}
	}
	panic(err)
}

// Default returns the embedded permissions.yml.
func Default() *Permissions {
	return embedded
}

// Paths returns every path a policy may set, in file order: each
// section's rules, All included, then the permissions.
func (p *Permissions) Paths() []string {
	var paths []string
	for _, section := range p.Sections {
		for _, rule := range p.Rules {
			paths = append(paths, section.Key+"/"+rule.Key)
		}
	}
	for _, permission := range p.Permissions {
		paths = append(paths, permission.Key)
	}
	return paths
}

// Checks returns the paths a check names: Paths without the All rules.
func (p *Permissions) Checks() []string {
	var paths []string
	for _, path := range p.Paths() {
		if _, rule, ok := strings.Cut(path, "/"); !ok || rule != All {
			paths = append(paths, path)
		}
	}
	return paths
}

// Valid reports whether a policy may set a path.
func (p *Permissions) Valid(path string) bool {
	section, rule, nested := strings.Cut(path, "/")
	if !nested {
		return has(p.Permissions, path)
	}
	return has(p.Sections, section) && has(p.Rules, rule)
}

// Checkable reports whether a check may name a path: valid, and not an
// All rule.
func (p *Permissions) Checkable(path string) bool {
	return p.Valid(path) && !strings.HasSuffix(path, "/"+All)
}

// ErrUnknown reports a path permissions.yml does not hold.
var ErrUnknown = errors.New("no such permission")

func has(list []Entry, key string) bool {
	for _, entry := range list {
		if entry.Key == key {
			return true
		}
	}
	return false
}
