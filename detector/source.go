package detector

// Source is one convention found in a folder: the file the detector
// read, the command that runs its targets, and the targets.
type Source struct {
	// Kind names the convention: atkins, task, make, npm, composer, script.
	Kind string `json:"kind"`
	// File is the file holding the targets, relative to the folder.
	File string `json:"file"`
	// Name and Description are what the file says about itself, when the
	// convention has a place for that.
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// Command is the argv prefix that runs a target, such as
	// `atkins -f atkins.yml` or `npm run`. A target's Args follow it.
	Command []string `json:"command"`

	// Includes are the files the source reads into its variables.
	Includes []string `json:"includes,omitempty"`
	// EnvFiles are the dotenv files the source reads, relative to the folder.
	EnvFiles []string `json:"env_files,omitempty"`

	Targets []*Target `json:"targets"`

	// Error holds the parse error when the file does not parse.
	Error string `json:"error,omitempty"`

	// Aside marks a source listed beside the one that reads the folder,
	// whose targets keep their names: the loose scripts of a scripts
	// folder with a pipeline, none named as a pipeline target.
	Aside bool `json:"aside,omitempty"`
}

// Target returns the target by name, or nil.
func (s *Source) Target(name string) *Target {
	if s == nil {
		return nil
	}
	for _, target := range s.Targets {
		if target.Name == name {
			return target
		}
	}
	return nil
}
