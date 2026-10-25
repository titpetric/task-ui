package integration

import (
	"unicode"
)

// Kind is an integration task-ui knows: its purpose, the settings it
// takes, and the steps to the credential among them.
type Kind struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Fields      []Field `json:"fields"`
	// Steps are how to get the integration's credential, in order, each
	// a sentence or two with `code` for commands and values.
	Steps []string `json:"steps"`
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
			Steps: []string{
				"On a machine where you are signed in to Claude Code, run `claude setup-token`.",
				"Sign in in the browser it opens. It prints a token starting `sk-ant-oat01-`, valid for a year.",
				"Paste it as the OAuth token here.",
				"Under Who, check the roles and users whose runs get it: the admin role, and the agent's user, `claude` on most instances, which the agent job of the Notes board runs as. Nothing checked means every run gets it.",
				"Save. A run of a checked user gets the token as `CLAUDE_CODE_OAUTH_TOKEN`. Set a new token before the year ends.",
			},
		},
		{
			ID:          Forgejo,
			Title:       "Forgejo",
			Description: "The git host. Sync makes a workspace of every organisation the token's user belongs to and of every user whose repositories it can read, with the repository named like the owner as its home and every other repository a project in its src, and installs a push hook on each the token administers so a push pulls the checkout. A user without a readable repository gets no workspace. It adds and updates, and never removes.",
			Fields: []Field{
				{Name: "url", Label: "URL", Placeholder: "https://git.cdn.si", Help: "The instance task-ui syncs from."},
				{Name: "token", Label: "Token", Secret: true, Help: "A personal access token of a user who is a member of the organisations to sync, with the scopes read:user, read:organization, read:repository and write:repository, the last for the hooks; see the Integrations chapter."},
				{Name: "clone", Label: "Clone over", Optional: true, Options: []string{"ssh", "https"}, Help: "ssh clones with the instance keys; https needs public repositories, as task-ui puts no token in a URL."},
				{Name: "secret", Label: "Webhook secret", Secret: true, Optional: true, Generated: true, Help: "Signs what Forgejo posts; empty makes one on the first save and keeps it."},
			},
			Steps: []string{
				"Sign in to Forgejo as a user who is a member of the organisations to sync.",
				"Under Settings, Applications, Generate token, make a token with the scopes `read:user`, `read:organization`, `read:repository` and `write:repository`; the first lists the users, the last is for the push hooks alone.",
				"Enter the instance URL and the token. Keep Clone over at `ssh`, which clones with the instance's git key.",
				"Leave the webhook secret empty: the first save makes one and keeps it.",
				"Save, then Dry run to read what a sync would do, then Sync now: a workspace per organisation, then one per user with a repository the token reads, `<login>/<login>` its home and every other repository a project.",
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
			Steps: []string{
				"In Beszel, add a read-only user that nobody signs in as, with a password nobody keeps.",
				"As a superuser, sign in to the hub's API: `POST /api/collections/_superusers/auth-with-password` with your identity and password answers with your token.",
				"Mint a token for the read-only user with it: `POST /api/collections/users/impersonate/<user id>` with `{\"duration\": 31536000}`, the seconds it lasts; the guide has both commands as curl.",
				"Enter the hub's URL as task-ui reaches it, and the token.",
				"Save. The Inventory page lists the hub's systems. Changing that user's password revokes the token.",
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

// Credential names what the steps lead to, for their heading: the label
// of the first secret the user brings, "token" or "OAuth token"; "" for a
// kind without one.
func (k *Kind) Credential() string {
	for _, field := range k.Fields {
		if !field.Secret || field.Generated {
			continue
		}
		label := []rune(field.Label)
		if len(label) > 1 && unicode.IsUpper(label[1]) {
			return field.Label
		}
		return string(unicode.ToLower(label[0])) + string(label[1:])
	}
	return ""
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
