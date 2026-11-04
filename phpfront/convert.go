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
//
// The converters are split by the page they answer, the way the host's
// bindings are: a folder's in convert_folder.go, the Inventory's in
// convert_inventory.go, the tooling statistics' in convert_tooling.go and the
// settings' in convert_settings.go. What is left here is what every page
// reaches: the three primitives, the chrome, the live runs and the two
// measures a folder and a tree share.

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
		out.Append(phpMap(
			"id", f.ID,
			"title", f.Title,
			"dir", f.Dir,
			"allow", texts(f.Allow),
			"sources", texts(f.Sources),
			"error", f.Error,
			"errors", findings(f.Errors),
			"jobs", jobs(f.Jobs),
		))
	}
	return out
}

// findings is what atkins' linter said about a folder's pipelines as a PHP
// array.
func findings(found []Finding) *model.Array {
	out := model.NewArraySize(len(found))
	for _, finding := range found {
		out.Append(phpMap("path", finding.Path, "job", finding.Job, "issue", finding.Issue, "detail", finding.Detail))
	}
	return out
}

// jobs is the jobs a folder declares as a PHP array. The folder's own allow
// list is the folder's key and not repeated under each of its jobs, so a page
// listing a catalogue reads the pair the gate takes from the two levels.
func jobs(found []Job) *model.Array {
	out := model.NewArraySize(len(found))
	for _, j := range found {
		out.Append(phpMap("id", j.ID, "name", j.Name, "dir", j.Dir, "shell", j.Shell, "host", j.Host, "allow", texts(j.Allow)))
	}
	return out
}

// job is one job as the catalogue resolved it, the whole gate in one answer:
// whether it is there at all, the two kinds a permission gates, and both allow
// lists, because the page that gates a run holds neither the folder nor the
// directory inside it the run named.
func job(found Job) *model.Array {
	return phpMap(
		"found", found.Found,
		"id", found.ID,
		"name", found.Name,
		"dir", found.Dir,
		"shell", found.Shell,
		"host", found.Host,
		"allow", texts(found.Allow),
		"folder_allow", texts(found.FolderAllow),
		"kind", found.Kind,
		"source", found.Source,
		"source_lines", found.SourceLines,
		"description", found.Description,
		"summary", found.Summary,
		"command", texts(found.Command),
		"steps", texts(found.Steps),
		"aliases", texts(found.Aliases),
		"depends", depends(found.Depends),
		"interactive", found.Interactive,
		"takes_args", found.TakesArgs,
		"args_command", found.ArgsCommand,
		"key", found.Key,
	)
}

// depends is the targets a job runs first as a PHP array.
func depends(found []Depends) *model.Array {
	out := model.NewArraySize(len(found))
	for _, dep := range found {
		out.Append(phpMap("text", dep.Text, "found", dep.Found, "id", dep.ID, "name", dep.Name, "dir", dep.Dir))
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

// flows is the workflow runs in progress as a PHP array.
func flows(found []Flow) *model.Array {
	out := model.NewArraySize(len(found))
	for _, f := range found {
		out.Append(phpMap("workflow_id", f.WorkflowID, "run_id", f.RunID))
	}
	return out
}

// session is what the runner holds about a run as a PHP array.
func session(held Session) *model.Array {
	return phpMap(
		"live", held.Live,
		"attached", held.Attached,
		"message", held.Message,
		"state", held.State,
		"exit_code", held.ExitCode,
		"title", held.Title,
		"finished_at", held.FinishedAt,
	)
}

// screen is the replayed last frame of a run as a PHP array.
func screen(shown Screen) *model.Array {
	return phpMap(
		"svg", shown.SVG,
		"lines", shown.Lines,
		"complete", shown.Complete,
	)
}

// rules is the permission file as a PHP array.
func rules(file Rules) *model.Array {
	return phpMap(
		"rules", titles(file.Rules),
		"sections", titles(file.Sections),
		"permissions", titles(file.Permissions),
		"paths", texts(file.Paths),
		"checks", texts(file.Checks),
		"extras", titles(file.Extras),
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

// tree is one directory of the artifact tree as a PHP array.
func treeItems(items []TreeItem) *model.Array {
	out := model.NewArraySize(len(items))
	for _, item := range items {
		out.Append(phpMap("name", item.Name, "dir", item.Dir, "size", item.Size, "mod_time", item.ModTime))
	}
	return out
}

// space is what a tree takes on disk as a PHP array.
func space(usage Space) *model.Array {
	return phpMap("measured", usage.Measured, "files", usage.Files, "history", usage.History, "ignored", usage.Ignored)
}

// spaces is what each of a page's directories takes as a PHP array, by the
// directory itself.
func spaces(found map[string]Space) *model.Array {
	out := model.NewArraySize(len(found))
	for dir, usage := range found {
		out.Set(dir, space(usage))
	}
	return out
}

// dirs is the directory of each of a tree's folders as a PHP array, by the
// folder ID. An ID the layout places nowhere is not a key, so a reader that
// asked about one tells that from a folder whose directory is the empty
// string, which no folder has.
func dirs(found map[string]string) *model.Array {
	out := model.NewArraySize(len(found))
	for id, dir := range found {
		out.Set(id, dir)
	}
	return out
}
