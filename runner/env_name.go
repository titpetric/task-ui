package runner

// EnvName is a variable of the server's environment as the Instance
// page lists it: its name, never its value, whether the name looks
// like a credential's, and whether a run gets it under EnvNone and
// EnvAuto; EnvAll passes every one.
type EnvName struct {
	Name      string `json:"name"`
	Sensitive bool   `json:"sensitive"`
	None      bool   `json:"none"`
	Auto      bool   `json:"auto"`
}

// Forwarded reports whether a run gets the variable under a mode.
func (n EnvName) Forwarded(mode string) bool {
	switch mode {
	case EnvAll:
		return true
	case EnvAuto:
		return n.Auto
	}
	return n.None
}
