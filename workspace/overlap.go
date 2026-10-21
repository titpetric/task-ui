package workspace

// Overlap is a key held for both uses: as an encryption key in Encrypt
// and as an SSH key in Access, by the paths of its files. task-ui
// treats it as an encryption key and withholds it from runs and git;
// an admin removes one of the copies.
type Overlap struct {
	Fingerprint string   `json:"fingerprint"`
	Encrypt     []string `json:"encrypt"`
	Access      []string `json:"access"`
}
