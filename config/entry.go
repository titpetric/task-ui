package config

// Entry is a key of permissions.yml with its title, in file order.
type Entry struct {
	Key   string `json:"key"`
	Title string `json:"title"`
}
