package storage

// ThreadUsage is what one thread read and wrote, and how many tools it called.
type ThreadUsage struct {
	Tools  int     `json:"tools"`
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}
