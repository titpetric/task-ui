package phpfront

import (
	"github.com/titpetric/phpscript/model"
)

// place is a folder, or a directory inside one, as a PHP array.
func place(found Place) *model.Array {
	return phpMap(
		"found", found.Found,
		"id", found.ID,
		"title", found.Title,
		"description", found.Description,
		"dir", found.Dir,
		"rel", found.Rel,
		"allow", texts(found.Allow),
		"hours", found.Hours,
		"error", found.Error,
		"errors", findings(found.Errors),
		"sources", sources(found.Sources),
		"includes", files(found.Includes),
	)
}

// sources is the pipeline files a place was detected in as a PHP array.
func sources(found []Source) *model.Array {
	out := model.NewArraySize(len(found))
	for _, source := range found {
		out.Append(phpMap(
			"kind", source.Kind,
			"file", source.File,
			"command", texts(source.Command),
			"lines", source.Lines,
		))
	}
	return out
}

// files is a list of files with the lines each holds as a PHP array.
func files(found []File) *model.Array {
	out := model.NewArraySize(len(found))
	for _, file := range found {
		out.Append(phpMap("name", file.Name, "lines", file.Lines))
	}
	return out
}

// groups is the Jobs table of a place as a PHP array.
func groups(found []Group) *model.Array {
	out := model.NewArraySize(len(found))
	for _, g := range found {
		out.Append(phpMap(
			"kind", g.Kind,
			"note", g.Note,
			"source", g.Source,
			"dir", g.Dir,
			"empty", g.Empty,
			"targets", g.Targets,
			"folder", where(g.Folder),
			"jobs", targets(g.Jobs),
		))
	}
	return out
}

// where is the folder a row's jobs run as as a PHP array. The keys are the
// ones Task\Folders::allows reads of a folder, so a row's gate is the folder's
// gate and not a second rule.
func where(found Where) *model.Array {
	return phpMap("id", found.ID, "rel", found.Rel, "allow", texts(found.Allow), "hours", found.Hours)
}

// targets is the jobs of one row of the Jobs table as a PHP array: what a
// button says, what it starts and what gates it, and none of the detail only
// the job's own page draws.
func targets(found []Job) *model.Array {
	out := model.NewArraySize(len(found))
	for _, j := range found {
		out.Append(phpMap(
			"id", j.ID,
			"name", j.Name,
			"dir", j.Dir,
			"kind", j.Kind,
			"source", j.Source,
			"description", j.Description,
			"key", j.Key,
			"interactive", j.Interactive,
			"takes_args", j.TakesArgs,
			"args_command", j.ArgsCommand,
			"shell", j.Shell,
			"host", j.Host,
			"allow", texts(j.Allow),
			"confirm", j.Confirm,
		))
	}
	return out
}

// entries is a directory the Files panel lists as a PHP array.
func entries(items []Entry) *model.Array {
	out := model.NewArraySize(len(items))
	for _, item := range items {
		out.Append(phpMap(
			"name", item.Name,
			"dir", item.Dir,
			"size", item.Size,
			"mod_time", item.ModTime,
			"path", item.Path,
			"folder", item.Folder,
			"sealed", item.Sealed,
			"exec", item.Exec,
		))
	}
	return out
}

// status is what git reports of a checkout as a PHP array.
func status(found Status) *model.Array {
	return phpMap(
		"found", found.Found,
		"checkout", found.Checkout,
		"branch", found.Branch,
		"detached", found.Detached,
		"upstream", found.Upstream,
		"ahead", found.Ahead,
		"behind", found.Behind,
		"dirty", found.Dirty,
		"changes", changes(found.Changes),
	)
}

// changes is the files a checkout has changed as a PHP array.
func changes(found []Change) *model.Array {
	out := model.NewArraySize(len(found))
	for _, change := range found {
		out.Append(phpMap("code", change.Code, "state", change.State, "path", change.Path, "from", change.From))
	}
	return out
}

// checkout is the commit a folder is at and the space it takes as a PHP array.
func checkout(found Checkout) *model.Array {
	return phpMap(
		"found", found.Found,
		"remote", found.Remote,
		"branch", found.Branch,
		"commit", found.Commit,
		"subject", found.Subject,
		"when", found.When,
		"space", space(found.Space),
	)
}

// hours is an hours rule read as a PHP array.
func hours(read Hours) *model.Array {
	return phpMap("line", read.Line, "open", read.Open)
}
