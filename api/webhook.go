package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/titpetric/task-ui/forgejo"
	"github.com/titpetric/task-ui/integration"
	"github.com/titpetric/task-ui/layout"
	"github.com/titpetric/task-ui/model"
)

// The errors the webhook answers with.
var (
	errWebhookMethod = errors.New("POST a hook")
	errWebhookOff    = errors.New("the Forgejo integration is not set up")
	errWebhookSigned = errors.New("the signature does not match the webhook secret")
	errWebhookBody   = errors.New("the hook's body")
	errWebhookPull   = errors.New("the pull failed")
)

// maxHookBody is the most of a hook's body the webhook reads.
const maxHookBody = 1 << 20

// Webhook answers POST /api/webhooks/forgejo, what the push hook sync
// installs on each repository calls: no session signs it in, the
// HMAC-SHA256 of the body under the Forgejo integration's webhook
// secret does, and it only ever pulls. A push pulls the project or the
// workspace home whose clone URL names the repository, by owner and
// name, under the lock a run pulls with; a ping answers so Forgejo's
// test button works. 404 with the integration off, 401 without the
// signature.
type Webhook struct {
	// Settings reads the Forgejo integration's settings.
	Settings func(ctx context.Context) (map[string]string, error)
	// Projects and Workspaces list the rows whose clone URLs name the
	// checkouts.
	Projects   func(ctx context.Context) ([]model.Project, error)
	Workspaces func(ctx context.Context) ([]model.Workspace, error)
	// Root holds the workspaces.
	Root string
	// Pull fast forwards a checkout under its lock.
	Pull func(ctx context.Context, dir string, w io.Writer) error
}

// ServeHTTP answers 202 with the folders pulled, 200 for a ping, a repository with no checkout or an event it ignores, else the error: 400 body, 401 signature, 404 integration off, 405 method, 502 pull.
func (h *Webhook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := h.serveHTTP(w, r); err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, errWebhookMethod):
			status = http.StatusMethodNotAllowed
		case errors.Is(err, errWebhookBody):
			status = http.StatusBadRequest
		case errors.Is(err, errWebhookSigned):
			status = http.StatusUnauthorized
		case errors.Is(err, errWebhookOff):
			status = http.StatusNotFound
		case errors.Is(err, errWebhookPull):
			status = http.StatusBadGateway
		}
		fail(w, status, err.Error())
	}
}

// serveHTTP checks the signature and handles the event, or returns why it does not.
func (h *Webhook) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return errWebhookMethod
	}
	values, err := h.Settings(r.Context())
	if err != nil {
		return err
	}
	if !integration.Configured(*integration.KindOf(integration.Forgejo), values) || values["secret"] == "" {
		return errWebhookOff
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, maxHookBody))
	if err != nil {
		return fmt.Errorf("%w: %w", errWebhookBody, err)
	}
	signature := r.Header.Get("X-Forgejo-Signature")
	if signature == "" {
		signature = r.Header.Get("X-Gitea-Signature")
	}
	if !forgejo.Verify(values["secret"], body, signature) {
		return errWebhookSigned
	}
	event := r.Header.Get("X-Forgejo-Event")
	if event == "" {
		event = r.Header.Get("X-Gitea-Event")
	}
	switch event {
	case "ping":
		writeJSON(w, http.StatusOK, map[string]string{"message": "pong"})
		return nil
	case "push":
		return h.push(w, r, body)
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ignored: " + event})
	return nil
}

// push pulls every checkout of the repository pushed, or returns why it did not.
func (h *Webhook) push(w http.ResponseWriter, r *http.Request, body []byte) error {
	push, err := forgejo.ParsePush(body)
	if err != nil {
		return fmt.Errorf("%w: %w", errWebhookBody, err)
	}
	dirs, err := h.checkouts(r.Context(), push.Key())
	if err != nil {
		return err
	}
	if len(dirs) == 0 {
		slog.Info("forgejo push for no checkout", "repository", push.Repository.FullName)
		writeJSON(w, http.StatusOK, map[string]string{"message": "no such checkout: " + push.Repository.FullName})
		return nil
	}
	var out bytes.Buffer
	var pulled []string
	for folder, dir := range dirs {
		if err := h.Pull(r.Context(), dir, &out); err != nil {
			slog.Warn("forgejo push pull failed", "repository", push.Repository.FullName, "folder", folder, "err", err)
			return fmt.Errorf("%w: %s: %w: %s", errWebhookPull, folder, err, strings.TrimSpace(out.String()))
		}
		pulled = append(pulled, folder)
	}
	slog.Info("forgejo push pulled", "repository", push.Repository.FullName, "ref", push.Ref, "folders", strings.Join(pulled, ", "))
	writeJSON(w, http.StatusAccepted, map[string]any{"pulled": sorted(pulled), "output": out.String()})
	return nil
}

// checkouts returns the directories of the checkouts of a repository,
// by folder: the projects whose source is it, and the homes of the
// workspaces whose home repository is it.
func (h *Webhook) checkouts(ctx context.Context, key string) (map[string]string, error) {
	dirs := map[string]string{}
	if key == "" {
		return dirs, nil
	}
	workspaces, err := h.Workspaces(ctx)
	if err != nil {
		return nil, err
	}
	var named []string
	for _, workspace := range workspaces {
		if workspace.Folder != model.Default {
			named = append(named, workspace.Folder)
		}
		if forgejo.Key(workspace.HomeURL) == key {
			dirs[layout.HomeID(workspace.Folder)] = layout.HomeDir(h.Root, workspace.Folder)
		}
	}
	projects, err := h.Projects(ctx)
	if err != nil {
		return nil, err
	}
	for _, project := range projects {
		if forgejo.Key(project.SourceURL) == key {
			dirs[project.Folder] = layout.Dir(h.Root, named, project.Folder)
		}
	}
	return dirs, nil
}
