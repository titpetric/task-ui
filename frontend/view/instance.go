package view

// Instance is what the server runs with, as the Instance step of the
// settings shows it: the configuration from the environment and the
// command line, less anything secret. A path shows; a token never does.
type Instance struct {
	// URL is TASK_UI_URL, empty when the links derive from each request.
	URL      string
	Listen   string
	Root     string
	DataDir  string
	SSHDir   string
	Htpasswd string
	Discover bool
	Admins   []string
	Viewers  []string
	// AutoLogin is the user every request signs in as, if any.
	AutoLogin string
	// Agent is the user a note names to run AgentJob on it.
	Agent    string
	AgentJob string
	// Beszel says the Beszel integration has a hub and a token, which
	// the Inventory reads.
	Beszel bool
}
