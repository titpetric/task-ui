package forgejo

import (
	"strings"
)

// Repo is a repository of an organisation, as the API lists them
// under an organisation.
type Repo struct {
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	DefaultBranch string `json:"default_branch"`
	SSHURL        string `json:"ssh_url"`
	CloneURL      string `json:"clone_url"`
	Archived      bool   `json:"archived"`
}

// Owner returns the owner of the repository, from its full name.
func (r Repo) Owner() string {
	owner, _, _ := strings.Cut(r.FullName, "/")
	return owner
}

// URL returns the URL a clone takes: the ssh one, or with https the
// https one.
func (r Repo) URL(https bool) string {
	if https {
		return r.CloneURL
	}
	return r.SSHURL
}

// Key returns the owner/name of a clone URL in lower case, so URLs of different schemes and hosts compare:
// ssh://git@git.cdn.si:2222/malta/apps.git, https://git.cdn.si/malta/apps.git and git@git.cdn.si:malta/apps
// are all malta/apps; "" for a URL without an owner and a name.
func Key(url string) string {
	url = strings.TrimSuffix(strings.TrimRight(strings.TrimSpace(url), "/"), ".git")
	i := strings.LastIndex(url, "/")
	if i < 0 {
		return ""
	}
	name, rest := url[i+1:], url[:i]
	owner := rest[strings.LastIndexAny(rest, "/:")+1:]
	if owner == "" || name == "" || strings.Contains(owner, "@") {
		return ""
	}
	return strings.ToLower(owner + "/" + name)
}
