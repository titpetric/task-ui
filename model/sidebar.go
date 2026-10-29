package model

// The kinds of a sidebar row. A view belongs to the instance and its row
// is never deleted; a module's row goes when the directory behind it does.
const (
	SidebarView   = "view"
	SidebarModule = "module"
)

// RuleSettings is the rule of a row that opens the settings. It is not a
// permission of config/permissions.yml: the settings open for whoever
// holds the view rule of any of its sections, which is what Page's
// CanSettings answers, and no single rule says that.
//
// Every other rule is a permission path, `module/view`, checked the way a
// template's `.Can` checks one. An empty rule is a row everyone signed in
// sees.
const RuleSettings = "settings"

// SidebarDefaults is the sidebar's top block as it stood before the table
// held it: the views across all folders, in this order, which is the
// order a fresh instance starts with.
//
// An icon names a symbol of the sprite in _layout.tpl without the i-
// prefix. The positions are the indexes, filled by EnsureDefaults.
func SidebarDefaults() []Sidebar {
	return []Sidebar{
		{Kind: SidebarView, Title: "Dashboard", Href: "/", Icon: "grid"},
		{Kind: SidebarView, Title: "Inventory", Href: "/inventory", Icon: "server"},
		{Kind: SidebarView, Title: "Workflows", Href: "/workflows", Icon: "flow"},
		{Kind: SidebarView, Title: "Runs", Href: "/runs", Icon: "list"},
		{Kind: SidebarView, Title: "Artifacts", Href: "/artifacts", Icon: "box"},
		{Kind: SidebarView, Title: "Scheduled", Href: "/schedules", Icon: "clock"},
		{Kind: SidebarView, Title: "Notes", Href: "/notes", Icon: "note"},
		{Kind: SidebarView, Title: "Settings", Href: "/settings", Icon: "settings", Rule: RuleSettings},
	}
}

// IsModule reports whether a row is a module's rather than the
// instance's, which is what decides whether a scan may drop it.
func (s *Sidebar) IsModule() bool { return s.Kind == SidebarModule }
