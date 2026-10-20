package api

// IntegrationRequest saves an integration's settings by name; an empty
// secret keeps the value it has.
type IntegrationRequest struct {
	Values map[string]string `json:"values"`
}
