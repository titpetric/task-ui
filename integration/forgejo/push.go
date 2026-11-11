package forgejo

import (
	"encoding/json"
	"errors"
)

// ErrNoRepository reports a hook body that names no repository.
var ErrNoRepository = errors.New("the hook names no repository")

// Push is what a push hook posts, the part task-ui reads: the ref
// pushed and the repository, whose clone URLs name the checkout to
// pull.
type Push struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
		SSHURL   string `json:"ssh_url"`
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
}

// ParsePush reads a push hook's body.
func ParsePush(body []byte) (*Push, error) {
	var push Push
	if err := json.Unmarshal(body, &push); err != nil {
		return nil, err
	}
	if push.Repository.FullName == "" && push.Repository.SSHURL == "" && push.Repository.CloneURL == "" {
		return nil, ErrNoRepository
	}
	return &push, nil
}

// Key returns the owner/name of the repository pushed, in lower case,
// to match a checkout's clone URL by.
func (p *Push) Key() string {
	for _, url := range []string{p.Repository.SSHURL, p.Repository.CloneURL, p.Repository.FullName} {
		if key := Key(url); key != "" {
			return key
		}
	}
	return ""
}
