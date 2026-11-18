package search

// Stats counts one index run.
type Stats struct {
	Documents   int
	Indexed     int
	Skipped     int
	Deleted     int
	Jobs        int
	JobsDeleted int
}
