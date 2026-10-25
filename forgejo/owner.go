package forgejo

// The kinds of owner a repository has.
const (
	OwnerOrg  = "org"
	OwnerUser = "user"
)

// Owner is who owns repositories on the instance, an organisation or a
// user, as sync sees them: both make a workspace by the same rules, the
// folder their lower case name, the repository named like them its
// home. Kind is OwnerOrg or OwnerUser.
type Owner struct {
	Name        string
	FullName    string
	Description string
	Kind        string
}

// Title returns the title of a workspace made of the owner: its full
// name, else its name.
func (o Owner) Title() string {
	if o.FullName != "" {
		return o.FullName
	}
	return o.Name
}
