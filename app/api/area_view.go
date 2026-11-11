package api

// AreaView is a section or an application permission of
// `config/permissions.yml`, with the sentence of `config/context.md` on
// its purpose, and what the user may do there: the rules of a section the
// user holds, and Allowed for any rule or for the permission.
type AreaView struct {
	Key     string   `json:"key"`
	Title   string   `json:"title"`
	About   string   `json:"about"`
	Rules   []string `json:"rules,omitempty"`
	Allowed bool     `json:"allowed"`
}
