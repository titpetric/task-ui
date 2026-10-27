package integration

import (
	"strings"
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
	Claude     = "claude"
	OpenRouter = "openrouter"
	Forgejo    = "forgejo"
	Beszel     = "beszel"
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
			Description: "The token Claude Code signs in with, from `claude setup-token`; a run gets it as CLAUDE_CODE_OAUTH_TOKEN for the users and roles checked below.",
			Fields: []Field{
				{Name: "token", Label: "OAuth token", Secret: true, Env: "CLAUDE_CODE_OAUTH_TOKEN", Placeholder: "sk-ant-oat01-…", Help: "Valid for a year from setup-token; set a new one before it ends."},
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
			ID:          OpenRouter,
			Title:       "OpenRouter",
			Description: "The gateway Claude Code asks instead of an Anthropic account; while set up, a run gets its variables and none of the Claude Code token.",
			Fields: []Field{
				{Name: "token", Label: "API key", Secret: true, Placeholder: "sk-or-v1-…", Help: "A key from openrouter.ai/settings/keys; it reaches a run as OPENROUTER_API_KEY and ANTHROPIC_AUTH_TOKEN."},
				{Name: "openrouter_api_key", Label: "OPENROUTER_API_KEY", Optional: true, Env: "OPENROUTER_API_KEY", Source: "token", Toggle: true, DefaultOn: true, Help: "OpenRouter's own name for the key, for the tools that read it."},
				{Name: "auth_token", Label: "ANTHROPIC_AUTH_TOKEN", Optional: true, Env: "ANTHROPIC_AUTH_TOKEN", Source: "token", Toggle: true, DefaultOn: true, Help: "What Claude Code signs in with; off falls back to the Claude Code integration."},
				{Name: "api_key", Label: "ANTHROPIC_API_KEY", Optional: true, Empty: true, Toggle: true, DefaultOn: true, Help: "Sent empty, so Claude Code signs in with the key alone."},
				{Name: "base_url", Label: "Base URL", Optional: true, Env: "ANTHROPIC_BASE_URL", Default: "https://openrouter.ai/api", Toggle: true, DefaultOn: true, Help: "Where Claude Code sends its requests."},
				{Name: "model", Label: "Model", Optional: true, Env: "ANTHROPIC_MODEL", Default: "deepseek/deepseek-v4.1-flash-1017", Toggle: true, DefaultOn: true, Help: "The model every Claude run asks for."},
				{Name: "max_context_tokens", Label: "Context tokens", Optional: true, Env: "CLAUDE_CODE_MAX_CONTEXT_TOKENS", Default: "1000000", Toggle: true, DefaultOn: true, Help: "The context window Claude Code plans for."},
				{Name: "auto_compact_window", Label: "Compact window", Optional: true, Env: "CLAUDE_CODE_AUTO_COMPACT_WINDOW", Default: "400000", Toggle: true, DefaultOn: true, Help: "How much of the context Claude Code fills before it compacts."},
				{Name: "gateway_model_discovery", Label: "Gateway model discovery", Optional: true, Env: "CLAUDE_CODE_ENABLE_GATEWAY_MODEL_DISCOVERY", Default: "1", Toggle: true, DefaultOn: true, Help: "Lets Claude Code list the models the gateway offers."},
				{Name: "disable_nonessential_traffic", Label: "Nonessential traffic", Optional: true, Env: "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC", Default: "1", Toggle: true, DefaultOn: true, Help: "Stops Claude Code's traffic that is not the conversation."},
				{Name: "enable_todo_tools", Label: "Todo tools", Optional: true, Env: "CLAUDE_CODE_ENABLE_TODO_TOOLS", Default: "1", Toggle: true, DefaultOn: true, Help: "Gives Claude Code its todo tools."},
			},
			Steps: []string{
				"Sign in at `https://openrouter.ai` and open Keys.",
				"Create a key named for this instance, and copy it: OpenRouter shows it once.",
				"The account needs credit, and the model must be one OpenRouter serves; the default is `deepseek/deepseek-v4.1-flash-1017`.",
				"Paste the key as the API key here, leave the switches on and Save.",
				"A Claude run then asks OpenRouter. Switch a variable off to leave it out of the runs; its value stays for the next switch on.",
			},
		},
		{
			ID:          Forgejo,
			Title:       "Forgejo",
			Description: "The git host; Sync makes a workspace of every organisation and of every user whose repositories the token reads, and installs a push hook on each repository it administers.",
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
			Description: "The host monitor; task-ui reads its systems for the Inventory page.",
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

// ToggleField returns the setting a request's "<name>.enabled" names, a
// setting with a switch; nil for any other name.
func (k *Kind) ToggleField(name string) *Field {
	base, ok := strings.CutSuffix(name, ".enabled")
	if !ok {
		return nil
	}
	field := k.Field(base)
	if field == nil || !field.Toggle {
		return nil
	}
	return field
}
