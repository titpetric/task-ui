package skills

// Finding is one problem atkins' linter found in a pipeline, as `atkins
// -l --json` prints it for a pipeline with findings: the pipeline's
// name and file, the job, the kind of issue and what is wrong.
type Finding struct {
	Pipeline string `json:"pipeline,omitempty"`
	Path     string `json:"path,omitempty"`
	Job      string `json:"job,omitempty"`
	Issue    string `json:"issue,omitempty"`
	Detail   string `json:"detail"`
}
