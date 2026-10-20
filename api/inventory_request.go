package api

// InventoryRequest runs a command on hosts of the inventory, by address;
// none is every host task-ui has an ssh login to.
type InventoryRequest struct {
	Command string   `json:"command"`
	Hosts   []string `json:"hosts"`
}
