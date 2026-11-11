package api

// WorkspaceExportRequest exports a workspace: the passphrase age
// encrypts the bundle to, which the import needs.
type WorkspaceExportRequest struct {
	Passphrase string `json:"passphrase"`
}
