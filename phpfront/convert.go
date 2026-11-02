package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

// A Go slice crosses into the runtime as a PHP array, and a Go struct crosses
// as a host object: its fields reach a script through published accessors and
// not through a key, so `$value["id"]` finds nothing and `get_object_vars`
// answers empty.
//
// Everything the host answers is read by a template, which indexes by key, and
// by the DAOs, which already speak arrays. So the boundary converts: what
// leaves here is an array all the way down, keyed by the names host.go
// declares as json tags.

// phpMap is a PHP array of alternating keys and values.
func phpMap(pairs ...any) *model.Array {
	out := model.NewArraySize(len(pairs) / 2)
	for i := 0; i+1 < len(pairs); i += 2 {
		out.Set(pairs[i], pairs[i+1])
	}
	return out
}

// phpList is a PHP array of values, keyed by position.
func phpList(values ...any) *model.Array {
	out := model.NewArraySize(len(values))
	for _, value := range values {
		out.Append(value)
	}
	return out
}

// strings is a list of strings as a PHP array.
func texts(values []string) *model.Array {
	out := model.NewArraySize(len(values))
	for _, value := range values {
		out.Append(value)
	}
	return out
}

// folders is the catalogue as a PHP array.
func folders(found []Folder) *model.Array {
	out := model.NewArraySize(len(found))
	for _, f := range found {
		out.Append(phpMap("id", f.ID, "title", f.Title, "dir", f.Dir, "jobs", texts(f.Jobs)))
	}
	return out
}

// running is the runs in progress as a PHP array.
func running(found []Running) *model.Array {
	out := model.NewArraySize(len(found))
	for _, r := range found {
		out.Append(phpMap("key", r.Key, "id", r.ID, "state", r.State))
	}
	return out
}

// rules is the permission file as a PHP array.
func rules(file Rules) *model.Array {
	return phpMap(
		"rules", titles(file.Rules),
		"sections", titles(file.Sections),
		"permissions", titles(file.Permissions),
		"paths", texts(file.Paths),
		"checks", texts(file.Checks),
		"extras", texts(file.Extras),
	)
}

// titles is a titled list of the permission file as a PHP array.
func titles(entries []Rule) *model.Array {
	out := model.NewArraySize(len(entries))
	for _, entry := range entries {
		out.Append(phpMap("key", entry.Key, "title", entry.Title))
	}
	return out
}
