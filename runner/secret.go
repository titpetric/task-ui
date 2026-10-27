package runner

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"

	"github.com/titpetric/task-ui/model"
)

// ErrRunToken refuses a token no live run holds.
var ErrRunToken = errors.New("no running run holds this token")

// ErrNoSecret answers a name no secret of the run's folder has.
var ErrNoSecret = errors.New("no secret of this name reaches the run's folder")

// runToken mints the token a run's task-ui get-secret signs in with,
// and returns the variables that carry it with the revoke that ends it
// once the run ends; without APIURL, none.
func (m *Manager) runToken(session *Session) ([]string, func(), error) {
	none := func() {}
	if m.opts.APIURL == nil {
		return nil, none, nil
	}
	url := m.opts.APIURL()
	if url == "" {
		return nil, none, nil
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return nil, none, err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	hash := hashToken(token)
	m.mu.Lock()
	m.runTokens[hash] = session
	m.mu.Unlock()
	revoke := func() {
		m.mu.Lock()
		delete(m.runTokens, hash)
		m.mu.Unlock()
	}
	return []string{"TASK_UI_API=" + url, "TASK_UI_RUN_TOKEN=" + token}, revoke, nil
}

// Secret returns the content of the secret of a name that reaches the
// folder of the run holding the token, as the run's environment has
// it, and records the read against the run. It fails with ErrRunToken
// for a token of no running run, and ErrNoSecret for a name the folder's
// secrets lack.
func (m *Manager) Secret(ctx context.Context, token, name string) (string, error) {
	m.mu.Lock()
	session, ok := m.runTokens[hashToken(token)]
	m.mu.Unlock()
	if !ok || token == "" {
		return "", ErrRunToken
	}
	if m.opts.Secrets == nil || session.noWorkspace {
		return "", ErrNoSecret
	}
	secrets, err := m.opts.Secrets.Reaching(ctx, session.Folder.ID)
	if err != nil {
		return "", err
	}
	for _, secret := range secrets {
		if secret.Name != name {
			continue
		}
		run := session.Run
		if m.opts.SecretReads != nil {
			read := model.SecretRead{SecretID: secret.ID, Name: secret.Name, RunID: run.ID, Folder: run.Folder, Dir: run.Dir, Job: run.Job, UserName: run.UserName}
			if err := m.opts.SecretReads.Record(ctx, []model.SecretRead{read}); err != nil {
				return "", err
			}
		}
		return secret.Content, nil
	}
	return "", ErrNoSecret
}

// hashToken returns the hex SHA-256 of a token, the key of runTokens.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
