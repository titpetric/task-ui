package forgejo

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// pageSize is how many rows one listing request takes at most.
const pageSize = 50

// ErrStatus reports an answer of the API that is no success.
var ErrStatus = errors.New("forgejo answered")

// Client calls the API v1 of a Forgejo instance with a personal access
// token: the organisations of the token's user, their repositories and
// the hooks of each. Every request has a timeout, and a listing
// follows the pages to the end.
type Client struct {
	// URL is the instance, such as https://git.cdn.si.
	URL   string
	Token string
	// HTTP sends the requests; nil is a client with a 10 second
	// timeout.
	HTTP *http.Client
}

// Orgs lists the organisations the token's user belongs to.
func (c *Client) Orgs(ctx context.Context) ([]Org, error) {
	var all []Org
	err := c.pages(ctx, "/api/v1/user/orgs", func(page []byte) (int, error) {
		var orgs []Org
		if err := json.Unmarshal(page, &orgs); err != nil {
			return 0, err
		}
		all = append(all, orgs...)
		return len(orgs), nil
	})
	return all, err
}

// Repos lists the repositories of an organisation, archived ones too.
func (c *Client) Repos(ctx context.Context, org string) ([]Repo, error) {
	var all []Repo
	err := c.pages(ctx, "/api/v1/orgs/"+url.PathEscape(org)+"/repos", func(page []byte) (int, error) {
		var repos []Repo
		if err := json.Unmarshal(page, &repos); err != nil {
			return 0, err
		}
		all = append(all, repos...)
		return len(repos), nil
	})
	return all, err
}

// Hooks lists the webhooks of a repository.
func (c *Client) Hooks(ctx context.Context, owner, repo string) ([]Hook, error) {
	var hooks []Hook
	err := c.do(ctx, http.MethodGet, c.hooksPath(owner, repo), nil, &hooks)
	return hooks, err
}

// CreateHook adds a webhook to a repository and returns it as Forgejo
// keeps it.
func (c *Client) CreateHook(ctx context.Context, owner, repo string, hook Hook) (*Hook, error) {
	var made Hook
	if err := c.do(ctx, http.MethodPost, c.hooksPath(owner, repo), hook, &made); err != nil {
		return nil, err
	}
	return &made, nil
}

// UpdateHook changes a webhook of a repository, by ID.
func (c *Client) UpdateHook(ctx context.Context, owner, repo string, id int64, hook Hook) (*Hook, error) {
	var changed Hook
	if err := c.do(ctx, http.MethodPatch, fmt.Sprintf("%s/%d", c.hooksPath(owner, repo), id), hook, &changed); err != nil {
		return nil, err
	}
	return &changed, nil
}

// hooksPath returns the path of a repository's hooks.
func (c *Client) hooksPath(owner, repo string) string {
	return "/api/v1/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/hooks"
}

// pages requests a listing page by page until a page comes back short,
// handing each page's body to read, which returns how many rows it held.
func (c *Client) pages(ctx context.Context, path string, read func(page []byte) (int, error)) error {
	for page := 1; ; page++ {
		var body json.RawMessage
		if err := c.do(ctx, http.MethodGet, fmt.Sprintf("%s?page=%d&limit=%d", path, page, pageSize), nil, &body); err != nil {
			return err
		}
		n, err := read(body)
		if err != nil {
			return err
		}
		if n < pageSize {
			return nil
		}
	}
}

// do sends one request with the token, in as JSON when not nil, and
// decodes the body into out; a status outside 2xx is ErrStatus with
// what the body said.
func (c *Client) do(ctx context.Context, method, path string, in, out any) error {
	var body io.Reader
	if in != nil {
		data, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(data)
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, strings.TrimRight(c.URL, "/")+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := c.HTTP
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, 8<<20))
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return fmt.Errorf("%w %d on %s %s: %s", ErrStatus, res.StatusCode, method, path, message(data))
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// message returns what an error answer said: its message field, else
// the body's first line.
func message(data []byte) string {
	var body struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(data, &body) == nil && body.Message != "" {
		return body.Message
	}
	line, _, _ := strings.Cut(strings.TrimSpace(string(data)), "\n")
	return line
}
