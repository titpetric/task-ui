package folder

import (
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/titpetric/task-ui/detector"
)

// FilePrefix starts the ID of a file job: ./bin/deploy is the job that
// runs the folder's bin/deploy.
const FilePrefix = "./"

// Path returns where a slash path inside the folder is on disk, with its
// symlinks resolved, and whether task-ui may open it: it exists, has no
// .. in it, is not sealed nor reached through .git, and stays inside
// the folder once Path follows its symlinks. Path holds where a symlink
// leads to the same rules: a link to a dot file opens no more than the
// dot file does. The empty path is the folder itself.
func (f *Folder) Path(rel string) (string, bool) {
	if f == nil || strings.HasPrefix(rel, "/") || strings.Contains(rel, "\\") {
		return "", false
	}
	for _, segment := range strings.Split(rel, "/") {
		if segment == ".." {
			return "", false
		}
	}
	if SealedPath(rel, true) {
		return "", false
	}
	root, err := filepath.EvalSymlinks(f.Dir)
	if err != nil {
		return "", false
	}
	resolved, err := filepath.EvalSymlinks(filepath.Join(f.Dir, filepath.FromSlash(rel)))
	if err != nil {
		return "", false
	}
	inner, err := filepath.Rel(root, resolved)
	if err != nil || inner == ".." || strings.HasPrefix(inner, ".."+string(filepath.Separator)) {
		return "", false
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", false
	}
	if SealedPath(rel, info.IsDir()) || (inner != "." && SealedPath(filepath.ToSlash(inner), info.IsDir())) {
		return "", false
	}
	return resolved, true
}

// FileJob returns the job that runs a file of the folder, rel to it, as
// detector.File reads it: ID ./<rel>, kind file, under the folder's
// allow rules and hours. Nil for a file Path refuses or one that does
// not run.
func (f *Folder) FileJob(rel string) *Job {
	full, ok := f.Path(rel)
	rel = path.Clean(rel)
	if !ok || rel == "." {
		return nil
	}
	if info, err := os.Stat(full); err != nil || !info.Mode().IsRegular() {
		return nil
	}
	target := detector.File(f.Dir, rel)
	if target == nil {
		return nil
	}
	return &Job{
		Target:  *target,
		Folder:  f.ID,
		ID:      FilePrefix + rel,
		Kind:    "file",
		Source:  rel,
		Command: target.Argv(),
	}
}

// Sub returns the folder of a directory inside this one, rel to it, as
// Load reads that directory: its own targets, each job with Dir set to
// rel, run in that directory under this folder's ID, allow rules and
// hours. The empty rel is the folder itself. Nil for a path Path
// refuses or one that is not a directory.
func (f *Folder) Sub(rel string) *Folder {
	if f == nil {
		return nil
	}
	rel = strings.Trim(rel, "/")
	if rel == "" {
		return f
	}
	if path.Clean(rel) != rel {
		return nil
	}
	dir, ok := f.Path(rel)
	if !ok {
		return nil
	}
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		return nil
	}
	sub := Load(f.ID, dir)
	sub.Title = f.ID + "/" + rel
	sub.Rel = rel
	sub.Allow, sub.Hours, sub.Probe = f.Allow, f.Hours, f.Probe
	for _, job := range sub.Jobs {
		job.Dir = rel
	}
	return sub
}
