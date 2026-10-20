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
			Description: "The token Claude Code signs in with, from `claude setup-token`, for the whole instance. A run gets it as CLAUDE_CODE_OAUTH_TOKEN when the user who started it, or their role, is checked below: the agent job of the Notes board runs as the agent's user, and a Claude terminal as you.",
			Fields: []Field{
				{Name: "token", Label: "OAuth token", Secret: true, Placeholder: "sk-ant-oat01-…", Help: "Valid for a year from setup-token; set a new one before it ends."},
				{Name: "who", Label: "Who", Optional: true, Who: true, Help: "The roles and users whose runs get the token; none checked is everyone."},
			},
		},
		{
			ID:          Beszel,
			Title:       "Beszel",
			Description: "The host monitor. task-ui reads its systems for the Inventory page: each host's state, cores, CPU, memory and disk. It signs in with a token, never a password.",
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
