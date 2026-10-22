package detector

// Target is an entry a source can run: a job of a pipeline, a target of a
// Makefile, a script.
type Target struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`
	Interactive bool     `json:"interactive,omitempty"`
	// TakesArgs says the target takes arguments from the command line,
	// after --: an atkins job with `args: true`. A run of any other
	// target takes none.
	TakesArgs bool `json:"takes_args,omitempty"`
	// Exec is the command the arguments of a target that takes them
	// follow, the argv of its exec: step, for a page to show what runs.
	// Empty when the target has none: the arguments follow its argv.
	Exec []string `json:"exec,omitempty"`

	// Args follow the source command: the target name for most
	// conventions, the script path for a script. Empty means the name.
	Args []string `json:"args,omitempty"`

	// Steps are the step labels, the way `atkins -l` prints them, or the
	// recipe lines of a make target, or the script of an npm target.
	Steps []string `json:"steps,omitempty"`
	// DependsOn names the targets that run first.
	DependsOn []string `json:"depends_on,omitempty"`
	// EnvFiles are the dotenv files this target reads on top of the source's.
	EnvFiles []string `json:"env_files,omitempty"`
}

// Argv returns what follows the source command: Args, or the name.
func (t *Target) Argv() []string {
	if len(t.Args) > 0 {
		return t.Args
	}
	return []string{t.Name}
}
