package forgejo

// User is a user of the instance, as GET /api/v1/users/search lists
// them.
type User struct {
	Login    string `json:"login"`
	FullName string `json:"full_name"`
}

// Owner returns the user as an owner of repositories.
func (u User) Owner() Owner {
	return Owner{Name: u.Login, FullName: u.FullName, Kind: OwnerUser}
}
