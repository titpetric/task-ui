package forgejo

// Org is an organisation of the instance, as GET /api/v1/user/orgs
// lists it.
type Org struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
}

// Owner returns the organisation as an owner of repositories.
func (o Org) Owner() Owner {
	return Owner{Name: o.Name, FullName: o.FullName, Description: o.Description, Kind: OwnerOrg}
}
