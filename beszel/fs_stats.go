package beszel

// FsStats is a filesystem an agent watches beside the root: its size and
// its use, in GB.
type FsStats struct {
	Total float64 `json:"d"`
	Used  float64 `json:"du"`
}
