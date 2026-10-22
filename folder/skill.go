package folder

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/titpetric/task-ui/detector"
	"github.com/titpetric/task-ui/skills"
)

// SkillPrefix starts the ID of a skill job: @git:pull is the job pull of
// the atkins skill git, as atkins finds it for the folder's directory.
const SkillPrefix = "@"

// Skill is an atkins skill that applies in a folder's directory, from a
// .atkins/skills folder of it or of a parent, with the jobs whose own
// condition holds there. Its jobs run `atkins <skill>:<job>` in Dir
// under the folder's allow rules and hours.
type Skill struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
	// Dir is where the jobs run, a slash path inside the folder: the
	// directory of the file the skill's condition matched. Empty is the
	// folder's own.
	Dir string `json:"dir"`
	// When is the condition, the matched file by its path inside the
	// folder, or by its name from outside.
	When *skills.When `json:"when,omitempty"`
	// Project is the compose project of the compose file the condition
	// matched: its top-level name, or its directory's name.
	Project string `json:"project,omitempty"`
	Jobs    []*Job `json:"jobs"`
}

// Skills returns the skills atkins applies in the folder's directory,
// in its order, each job one of the folder's. None without a Probe.
func (f *Folder) Skills() []*Skill {
	if f == nil || f.Probe == nil {
		return nil
	}
	var list []*Skill
	for _, section := range f.Probe.Probe(f.Dir) {
		if !section.IsSkill() {
			continue
		}
		skill := &Skill{ID: section.ID, Title: section.Desc, Dir: f.dirOf(section.Dir)}
		if section.When != nil {
			when := *section.When
			skill.Project = composeProject(when.Matched)
			when.Matched = f.shown(when.Matched)
			skill.When = &when
		}
		for _, cmd := range section.Cmds {
			if !cmd.Active {
				continue
			}
			skill.Jobs = append(skill.Jobs, f.skillJob(skill, section, cmd))
		}
		if len(skill.Jobs) > 0 {
			DefaultFirst(skill.Jobs)
			list = append(list, skill)
		}
	}
	return list
}

// SkillJob returns the job of a skill by its atkins ID, skill:job, as
// Skills lists it, or nil.
func (f *Folder) SkillJob(id string) *Job {
	for _, skill := range f.Skills() {
		for _, job := range skill.Jobs {
			if job.ID == SkillPrefix+id {
				return job
			}
		}
	}
	return nil
}

// skillJob makes a skill's job one of the folder's: ID @skill:job, kind
// skill, run as atkins skill:job in the skill's directory, with the
// atkins the probe ran.
func (f *Folder) skillJob(skill *Skill, section skills.Section, cmd skills.Command) *Job {
	job := &Job{
		Target: detector.Target{
			Name:        cmd.Name,
			Description: cmd.Desc,
			Interactive: cmd.Interactive,
			TakesArgs:   cmd.Args,
			Exec:        cmd.Exec,
			Steps:       []string{cmd.Cmd},
		},
		Folder:  f.ID,
		Dir:     skill.Dir,
		ID:      SkillPrefix + cmd.ID,
		Kind:    "skill",
		Source:  f.shown(section.Path),
		Command: []string{f.Probe.Bin(), cmd.ID},
	}
	if job.Name == "" {
		job.Name = cmd.ID
	}
	return job
}

// dirOf returns where a directory atkins named is inside the folder, a
// slash path from the folder's top: the folder's own directory, or one
// of its parents up to the top. A directory outside, or below, is the
// folder's own: atkins finds the skill from there again.
func (f *Folder) dirOf(dir string) string {
	here, err := filepath.EvalSymlinks(f.Dir)
	if err != nil || dir == "" {
		return f.Rel
	}
	up, err := filepath.Rel(dir, here)
	if err != nil {
		return f.Rel
	}
	if up == "." {
		return f.Rel
	}
	rel := f.Rel
	for _, segment := range strings.Split(filepath.ToSlash(up), "/") {
		if segment == ".." || rel == "" {
			return f.Rel
		}
		rel = path.Dir(rel)
		if rel == "." {
			rel = ""
		}
	}
	return rel
}

// shown returns a file atkins named as a page shows it: its path inside
// the folder's directory, or its name.
func (f *Folder) shown(file string) string {
	if file == "" {
		return ""
	}
	if here, err := filepath.EvalSymlinks(f.Dir); err == nil {
		if rel, err := filepath.Rel(here, file); err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.Base(file)
}

// composeName is the top-level name of a compose file.
var composeName = regexp.MustCompile(`(?m)^name:\s*["']?([A-Za-z0-9_.-]+)`)

// composeProject returns the project a compose file is: its top-level
// name, or its directory's name, as docker compose names it; "" for a
// file that is not a compose file.
func composeProject(file string) string {
	base := filepath.Base(file)
	if !strings.Contains(base, "compose.y") {
		return ""
	}
	if data, err := os.ReadFile(file); err == nil {
		if m := composeName.FindSubmatch(data); m != nil {
			return strings.ToLower(string(m[1]))
		}
	}
	return strings.ToLower(filepath.Base(filepath.Dir(file)))
}
