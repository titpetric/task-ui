// Package rulelint matches the permission checks of the code against
// `config/permissions.yml`: every rule a check names is one the configuration
// holds, and a check names every rule the configuration holds.
//
// A check is a call of a function or method named in Checks, with the
// rule its last argument, a string literal; in a template, Can or can
// followed by a quoted rule. A check function that passes its own
// argument to another is no finding.
package rulelint

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/titpetric/task-ui/config"
)

// Checks are the names of the functions and methods whose last argument
// is the rule they check.
var Checks = map[string]bool{"Allowed": true, "Can": true, "can": true, "require": true}

// templateCheck finds a check in a template: {{ if .Can "note/edit" }}.
var templateCheck = regexp.MustCompile(`\b[Cc]an "([^"]*)"`)

// Lint checks the Go files but tests, and the templates under root, against the configuration.
// It returns the mismatches sorted: unknown rules, rules of no literal, and rules nothing checks.
func Lint(root string, perms *config.Permissions) ([]string, error) {
	var findings []string
	used := map[string]bool{}
	use := func(at, rule string) {
		switch {
		case strings.HasSuffix(rule, "/"+config.All) && perms.Valid(rule):
			findings = append(findings, fmt.Sprintf("%s: rule %q is a policy's shorthand, no check names it", at, rule))
		case !perms.Checkable(rule):
			findings = append(findings, fmt.Sprintf("%s: unknown rule %q", at, rule))
		default:
			used[rule] = true
		}
	}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if path != root && (name == "testdata" || name == "vendor" || name == "node_modules" || strings.HasPrefix(name, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		switch {
		case strings.HasSuffix(name, "_test.go"):
			return nil
		case strings.HasSuffix(name, ".go"):
			return lintGo(path, rel, use, &findings)
		case strings.HasSuffix(name, ".tpl"):
			return lintTemplate(path, rel, use)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, rule := range perms.Checks() {
		if !used[rule] {
			findings = append(findings, fmt.Sprintf("config: rule %q never checked", rule))
		}
	}
	sort.Strings(findings)
	return findings, nil
}

// lintGo reports the checks of a Go file.
func lintGo(path, rel string, use func(at, rule string), findings *[]string) error {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return err
	}
	at := func(pos token.Pos) string {
		return fmt.Sprintf("%s:%d", filepath.ToSlash(rel), fset.Position(pos).Line)
	}
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		forwards := Checks[fn.Name.Name]
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || !Checks[calleeName(call.Fun)] || len(call.Args) == 0 {
				return true
			}
			arg := call.Args[len(call.Args)-1]
			lit, ok := arg.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				if !forwards {
					*findings = append(*findings, fmt.Sprintf("%s: the rule of %s is no string literal", at(arg.Pos()), calleeName(call.Fun)))
				}
				return true
			}
			rule, err := strconv.Unquote(lit.Value)
			if err != nil {
				rule = lit.Value
			}
			use(at(lit.Pos()), rule)
			return true
		})
	}
	return nil
}

// calleeName returns the name of the function or method a call calls.
func calleeName(fun ast.Expr) string {
	switch f := fun.(type) {
	case *ast.Ident:
		return f.Name
	case *ast.SelectorExpr:
		return f.Sel.Name
	}
	return ""
}

// lintTemplate reports the checks of a template.
func lintTemplate(path, rel string, use func(at, rule string)) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for i, line := range strings.Split(string(data), "\n") {
		for _, match := range templateCheck.FindAllStringSubmatch(line, -1) {
			use(fmt.Sprintf("%s:%d", filepath.ToSlash(rel), i+1), match[1])
		}
	}
	return nil
}
