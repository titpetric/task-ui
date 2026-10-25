package forgejo

// Org is an organisation of the instance, as GET /api/v1/user/orgs
// lists it.
type Org struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

// Title returns the title of a workspace made of the organisation:
// its full name, else its name.
func (o Org) Title() string {
	if o.FullName != "" {
		return o.FullName
	}
	return o.Name
}
