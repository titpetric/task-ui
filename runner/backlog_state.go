package runner

// BacklogState is how the slots and the line for them stand, as the
// Instance page and GET /api/instance show it.
type BacklogState struct {
	// Parallel is how many runs do work at once, the CPUs; Going how
	// many do now, and Waiting how many wait for a slot.
	Parallel int `json:"parallel"`
	Going    int `json:"going"`
	Waiting  int `json:"waiting"`
	// Auto is the automatic backlog length, BacklogPerCPU times the
	// CPUs; Configured the one an admin set, 0 for none; Length the one
	// in effect.
	Auto       int `json:"backlog_length_auto"`
	Configured int `json:"backlog_length"`
	Length     int `json:"backlog_length_effective"`
}
