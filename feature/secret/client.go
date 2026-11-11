package secret

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is task-ui get-secret's side of a run: it requests from the API
// of the instance that started the run, TASK_UI_API, a secret that
// reaches the run's folder, signed in with the run's own token,
// TASK_UI_RUN_TOKEN.
type Client struct {
	API   string
	Token string
	// HTTP sends the request, nil for http.DefaultClient.
	HTTP *http.Client
}

// ClientFromEnv returns the client of the run getenv reads; it fails
// without TASK_UI_API and TASK_UI_RUN_TOKEN, outside a run of task-ui.
func ClientFromEnv(getenv func(string) string) (*Client, error) {
	c := &Client{API: strings.TrimRight(getenv("TASK_UI_API"), "/"), Token: getenv("TASK_UI_RUN_TOKEN")}
	if c.API == "" || c.Token == "" {
		return nil, errors.New("TASK_UI_API and TASK_UI_RUN_TOKEN are unset: task-ui get-secret works in a run task-ui started")
	}
	return c, nil
}

// Get returns the content of the secret of a name. It fails with the
// server's message when the run ended, no secret of the name reaches
// the run's folder, or the request does not go through.
func (c *Client) Get(ctx context.Context, name string) (string, error) {
	if name == "" || strings.ContainsAny(name, "/?#") {
		return "", fmt.Errorf("%q is no secret name", name)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.API+"/run/secrets/"+url.PathEscape(name), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	client := c.HTTP
	if client == nil {
		client = http.DefaultClient
	}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var body struct {
		Value string `json:"value"`
		Error string `json:"error"`
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return "", fmt.Errorf("%s: %s", res.Status, strings.TrimSpace(string(data)))
	}
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", res.Status, body.Error)
	}
	return body.Value, nil
}
