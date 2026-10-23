package view

// The steps of the settings, by name, as a page's Step names its own.
const (
	StepInstance     = "Instance"
	StepRetention    = "Data retention"
	StepKeys         = "Keys"
	StepSecrets      = "Secrets"
	StepIntegrations = "Integrations"
	StepUsers        = "Users"
	StepPermissions  = "Permissions"
	StepWorkflows    = "Workflows"
	StepWorkspaces   = "Workspaces"
	StepProjects     = "Projects"
)

// SetupStep is a step of the settings in the order an admin sets up an
// instance: its page, and its place among the steps a user may open.
type SetupStep struct {
	Name string
	Href string
	Icon string
	// Desc says what the step's page holds, on the Settings page.
	Desc string
	// Number is the step's place among the steps the user may open,
	// from 1.
	Number int
	// Current is the step of the page that lists the steps.
	Current bool
}

// setupOrder is the order of the settings, the one list the Settings
// page and the footer of every step read: the instance first and how
// long it keeps what it records, then the keys and the secrets they encrypt, the integrations, the users and
// what they may do, the workflows, and the workspaces and projects
// they run in. Opens says whether a user may open the step's page.
var setupOrder = []struct {
	step  SetupStep
	opens func(p *Page) bool
}{
	{SetupStep{Name: StepInstance, Href: "/settings/instance", Icon: "settings", Desc: "What the server runs with: its link, version, folders, admins and viewers, and the agent job. Read only here: the environment sets it."}, (*Page).CanSettings},
	{SetupStep{Name: StepRetention, Href: "/settings/retention", Icon: "clock", Desc: "How long the instance keeps run recordings, earlier artifact versions and the secret read log, and how much of each; the oldest go first."}, (*Page).CanSettings},
	{SetupStep{Name: StepKeys, Href: "/settings/keys", Icon: "key", Desc: "The encryption key of each workspace, which stores its secrets, and the SSH keys that connect runs and git to hosts."}, func(p *Page) bool { return p.Can("key/view") }},
	{SetupStep{Name: StepSecrets, Href: "/settings/secrets", Icon: "secret", Desc: "Variables that runs in a workspace get in their environment, for every project in it or for one. Secrets need an encryption key."}, func(p *Page) bool { return p.Can("secret/view") }},
	{SetupStep{Name: StepIntegrations, Href: "/settings/integrations", Icon: "plug", Desc: "The services task-ui talks to, Claude Code and Beszel, with their API keys and tokens, encrypted with the instance key."}, func(p *Page) bool { return p.Can("integration/view") }},
	{SetupStep{Name: StepUsers, Href: "/settings/users", Icon: "users", Desc: "Who can sign in, with each user's role and folders."}, func(p *Page) bool { return p.Can("user/view") }},
	{SetupStep{Name: StepPermissions, Href: "/settings/permissions", Icon: "key", Desc: "What each role can do, rule by rule: inherit, allow or deny. Each user's profile holds their own settings."}, func(p *Page) bool { return p.Can("user/view") }},
	{SetupStep{Name: StepWorkflows, Href: "/workflows", Icon: "flow", Desc: "Chains of jobs that run one after another, started by hand, on a schedule or from a webhook."}, func(p *Page) bool { return p.Can("workflow/view") }},
	{SetupStep{Name: StepWorkspaces, Href: "/settings/workspaces", Icon: "folder-open", Desc: "Checkouts that hold projects: an organisation's tooling, with its projects cloned inside the workspace's folder."}, func(p *Page) bool { return p.Can("workspace/view") }},
	{SetupStep{Name: StepProjects, Href: "/settings/projects", Icon: "folder", Desc: "The folders under the root. Add one from a git repository; task-ui pulls the checkout before every run."}, func(p *Page) bool { return p.Can("project/view") }},
}

// SetupSteps returns the steps of the settings the user may open, in
// setup order and numbered, with the page's Step current.
func (p *Page) SetupSteps() []SetupStep {
	var steps []SetupStep
	for _, entry := range setupOrder {
		if !entry.opens(p) {
			continue
		}
		step := entry.step
		step.Number = len(steps) + 1
		step.Current = step.Name == p.Step
		steps = append(steps, step)
	}
	return steps
}
