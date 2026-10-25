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
	Claude  = "claude"
	Forgejo = "forgejo"
	Beszel  = "beszel"
)

// WebhookPath is the path of the Forgejo webhook under the instance's
// public URL.
const WebhookPath = "/api/webhooks/forgejo"

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
			ID:          Forgejo,
			Title:       "Forgejo",
			Description: "The git host. Sync makes a workspace of every organisation the token's user belongs to, with the repository named like the organisation as its home and every other repository a project in its src, and installs a push hook on each so a push pulls the checkout. It adds and updates, and never removes.",
			Fields: []Field{
				{Name: "url", Label: "URL", Placeholder: "https://git.cdn.si", Help: "The instance task-ui syncs from."},
				{Name: "token", Label: "Token", Secret: true, Help: "A personal access token of a user who is a member of the organisations to sync, with the scopes read:organization, read:repository and write:repository, the last for the hooks; see the Integrations chapter."},
				{Name: "clone", Label: "Clone over", Optional: true, Options: []string{"ssh", "https"}, Help: "ssh clones with the instance keys; https needs public repositories, as task-ui puts no token in a URL."},
				{Name: "secret", Label: "Webhook secret", Secret: true, Optional: true, Generated: true, Help: "Signs what Forgejo posts; empty makes one on the first save and keeps it."},
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
