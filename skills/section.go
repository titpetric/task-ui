package skills

// Section is one file of jobs atkins found for a directory: the pipeline
// of the directory, or a skill from a .atkins/skills folder of it or a
// parent. Dir is where its jobs run, the directory of the file a skill's
// condition matched; Active says the condition holds.
type Section struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Dir    string `json:"dir,omitempty"`
	Active bool   `json:"active"`
	When   *When  `json:"when,omitempty"`
	Desc   string `json:"desc,omitempty"`

	// Help is the pipeline's own account of its purpose, its help: block.
	Help string `json:"help,omitempty"`

	// Include are the files the section reads into its variables, and
	// Env the dotenv files it reads into the environment.
	Include []string `json:"include,omitempty"`
	Env     []string `json:"env,omitempty"`

	Cmds []Command `json:"cmds"`
}

// Kinds of section.
const (
	KindPipeline = "pipeline"
	KindSkill    = "skill"
)

// IsSkill reports whether the section is a skill that applies here.
func (s *Section) IsSkill() bool {
	return s.Kind == KindSkill && s.Active
}
