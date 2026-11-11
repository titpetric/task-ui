package spotify

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/titpetric/task-ui/app/model"
	"github.com/titpetric/task-ui/feature/secret"
	"github.com/titpetric/task-ui/integration"
)

// UserScope is the secret scope of a user's own variables, encrypted with
// the instance key like an integration's, and read by no run.
const UserScope = integration.UserScopePrefix

// Link is what task-ui keeps of a user's linked account.
type Link struct {
	Token
	// Name is the account's display name, for the page.
	Name string
	// Broken marks a link Spotify refused to refresh.
	Broken bool
}

// Tokens keeps the links.
type Tokens interface {
	// Load returns a user's link, nil when there is none.
	Load(ctx context.Context, user string) (*Link, error)
	Save(ctx context.Context, user string, link Link) error
	// Delete removes the link and every token it holds.
	Delete(ctx context.Context, user string) error
}

// SecretTokens keeps the links as secrets of the user's own scope.
type SecretTokens struct {
	Secrets *secret.Store
}

const (
	nameAccess  = "spotify_access"
	nameRefresh = "spotify_refresh"
	nameExpires = "spotify_expires"
	nameScope   = "spotify_scope"
	nameAccount = "spotify_account"
	nameBroken  = "spotify_broken"
)

// Load returns the user's link, nil without a refresh token.
func (t SecretTokens) Load(ctx context.Context, user string) (*Link, error) {
	rows, err := t.Secrets.Scope(ctx, UserScope+user)
	if err != nil {
		return nil, err
	}
	values := map[string]string{}
	for _, row := range rows {
		values[row.Name] = row.Content
	}
	if values[nameRefresh] == "" {
		return nil, nil
	}
	link := &Link{Name: values[nameAccount], Broken: values[nameBroken] == "1"}
	link.Access, link.Refresh, link.Scope = values[nameAccess], values[nameRefresh], values[nameScope]
	if unix, err := strconv.ParseInt(values[nameExpires], 10, 64); err == nil {
		link.Expires = time.Unix(unix, 0)
	}
	return link, nil
}

// Save writes the link, every value encrypted.
func (t SecretTokens) Save(ctx context.Context, user string, link Link) error {
	broken := "0"
	if link.Broken {
		broken = "1"
	}
	for name, value := range map[string]string{
		nameAccess:  link.Access,
		nameRefresh: link.Refresh,
		nameExpires: strconv.FormatInt(link.Expires.Unix(), 10),
		nameScope:   link.Scope,
		nameAccount: link.Name,
		nameBroken:  broken,
	} {
		if err := t.Secrets.Set(ctx, &model.Secret{Scope: UserScope + user, Name: name, Content: value}); err != nil {
			return err
		}
	}
	return nil
}

// Delete removes the user's Spotify secrets.
func (t SecretTokens) Delete(ctx context.Context, user string) error {
	rows, err := t.Secrets.Scope(ctx, UserScope+user)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if strings.HasPrefix(row.Name, "spotify_") {
			if err := t.Secrets.Delete(ctx, row.ID); err != nil {
				return err
			}
		}
	}
	return nil
}
