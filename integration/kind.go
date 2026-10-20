package integration

// Kind is an integration task-ui knows: its purpose and the settings
// it takes.
type Kind struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
}

// The integrations.
const (
	Claude = "claude"
	Beszel = "beszel"
)

// Kinds returns the integrations, in the order Settings lists them.
func Kinds() []Kind {
	return []Kind{
		{
			ID:          Claude,
			Title:       "Claude Code",
			Description: "The token Claude Code signs in with, from `claude setup-token`. Runs of the folders below get it as CLAUDE_CODE_OAUTH_TOKEN: the agent job of the Notes board, and a Claude terminal.",
			Fields: []Field{
				{Name: "token", Label: "OAuth token", Secret: true, Placeholder: "sk-ant-oat01-…", Help: "Valid for a year from setup-token; set a new one before it ends."},
				{Name: "folders", Label: "Folders", Placeholder: "deploy/scripts", Help: "The folders whose runs get the token, comma separated."},
			},
		},
		{
			ID:          Beszel,
			Title:       "Beszel",
			Description: "The host monitor. task-ui reads its systems for the Hosts page: each host's state, load, memory, disk and uptime. It signs in with a token, never a password.",
			Fields: []Field{
				{Name: "url", Label: "URL", Placeholder: "http://beszel:8090", Help: "Where task-ui reaches the hub."},
				{Name: "token", Label: "Token", Secret: true, Help: "An impersonation token of a read-only Beszel user, minted by a superuser with a duration; see the Integrations chapter. Changing that user's password revokes it."},
			},
		},
	}
}

// KindOf returns an integration by ID, or nil.
func KindOf(id string) *Kind {
	for _, kind := range Kinds() {
		if kind.ID == id {
			return &kind
		}
	}
	return nil
}

// Field returns a setting by name, or nil.
func (k *Kind) Field(name string) *Field {
	for i := range k.Fields {
		if k.Fields[i].Name == name {
			return &k.Fields[i]
		}
	}
	return nil
}
