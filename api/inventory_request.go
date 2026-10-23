package api

// InventoryRequest runs a command on hosts of the inventory, by address,
// a run of @host:<host> each under one batch; none is every host task-ui has
// an ssh login to. With Summarize, Claude summarizes the batch in a
// thread about its first run after every host finishes.
type InventoryRequest struct {
	Command   string   `json:"command"`
	Hosts     []string `json:"hosts"`
	Summarize bool     `json:"summarize,omitempty"`
}
