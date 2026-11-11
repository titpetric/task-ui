package form

// Submission is a bound submission: what each visible field holds, what a multi holds, and what is wrong.
type Submission struct {
	// Values hold one value per visible field, a multi joined with ", ".
	Values map[string]string
	// List holds the parts of each visible multi.
	List map[string][]string
	// Errors hold a message per field the submission got wrong.
	Errors map[string]string
}
