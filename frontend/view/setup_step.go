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
	{SetupStep{Name: StepInstance, Href: "/settings/instance", Icon: "settings", Desc: "The server's settings, from the environment."}, (*Page).CanSettings},
	{SetupStep{Name: StepRetention, Href: "/settings/retention", Icon: "clock", Desc: "How long runs, recordings, threads, artifact history and the read log stay."}, (*Page).CanSettings},
	{SetupStep{Name: StepKeys, Href: "/settings/keys", Icon: "key", Desc: "The encryption key that stores the secrets, and the SSH keys that connect runs and git to hosts."}, func(p *Page) bool { return p.Can("key/view") }},
	{SetupStep{Name: StepSecrets, Href: "/settings/secrets", Icon: "secret", Desc: "Run environment variables, per workspace or project."}, func(p *Page) bool { return p.Can("secret/view") }},
	{SetupStep{Name: StepIntegrations, Href: "/settings/integrations", Icon: "plug", Desc: "Claude Code and Beszel, and their tokens."}, func(p *Page) bool { return p.Can("integration/view") }},
	{SetupStep{Name: StepUsers, Href: "/settings/users", Icon: "users", Desc: "Who signs in, their role and folders."}, func(p *Page) bool { return p.Can("user/view") }},
	{SetupStep{Name: StepPermissions, Href: "/settings/permissions", Icon: "key", Desc: "What each role may do."}, func(p *Page) bool { return p.Can("user/view") }},
	{SetupStep{Name: StepWorkflows, Href: "/workflows", Icon: "flow", Desc: "Chains of jobs, by hand, schedule or webhook."}, func(p *Page) bool { return p.Can("workflow/view") }},
	{SetupStep{Name: StepWorkspaces, Href: "/settings/workspaces", Icon: "folder-open", Desc: "Directories that hold projects and their tooling."}, func(p *Page) bool { return p.Can("workspace/view") }},
	{SetupStep{Name: StepProjects, Href: "/settings/projects", Icon: "folder", Desc: "Git checkouts under the root, pulled before each run."}, func(p *Page) bool { return p.Can("project/view") }},
}

// SetupSteps returns the numbered steps of the settings the user may open, in order, the page's Step current.
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
