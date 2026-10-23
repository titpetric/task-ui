package skills

// Command is one job of a section, as `atkins -l --json` lists it: ID is
// what runs it, skill:job for a skill, Name the job as the file spells it.
type Command struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Desc        string `json:"desc,omitempty"`
	Cmd         string `json:"cmd"`
	Interactive bool   `json:"interactive"`
	// Args says the job takes arguments, after --.
	Args bool `json:"args"`
	// Exec is the argv of the first exec: step of a job that takes
	// args: the command the arguments follow.
	Exec []string `json:"exec,omitempty"`

	// Aliases are the other names the job answers to, and DependsOn the
	// jobs that run before it.
	Aliases   []string `json:"aliases,omitempty"`
	DependsOn []string `json:"depends_on,omitempty"`

	// Steps are the step labels, the lines a run prints under the job.
	Steps []string `json:"steps,omitempty"`

	// Env are the dotenv files the job reads on top of its section's.
	Env []string `json:"env,omitempty"`

	// Hidden reports a job the file asks not to be offered, with
	// show: false. It is listed so a reader knows it exists.
	Hidden bool `json:"hidden,omitempty"`

	Active bool  `json:"active"`
	When   *When `json:"when,omitempty"`
}
