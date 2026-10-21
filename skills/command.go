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
	Exec   []string `json:"exec,omitempty"`
	Active bool     `json:"active"`
	When   *When    `json:"when,omitempty"`
}
