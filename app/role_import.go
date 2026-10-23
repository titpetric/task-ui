package app

import (
	"context"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/titpetric/task-ui/model"
	"github.com/titpetric/task-ui/storage"
)

// RoleImport is the role lists of an older configuration: the admins,
// the viewers and the agent user, by name.
type RoleImport struct {
	Admins  []string
	Viewers []string
	Agent   string
}

// roleImport reads the role lists of an older configuration, or nil
// when the environment holds none of them.
func roleImport() *RoleImport {
	admins, viewers, agent := csv(os.Getenv("TASK_UI_ADMINS")), csv(os.Getenv("TASK_UI_VIEWERS")), strings.TrimSpace(os.Getenv("TASK_UI_AGENT"))
	if len(admins) == 0 && len(viewers) == 0 && agent == "" {
		return nil
	}
	return &RoleImport{Admins: admins, Viewers: viewers, Agent: agent}
}

// importedRoles is the setting that says the role lists of an older
// configuration went into storage: the environment is not read again.
const importedRoles = "users.roles_imported"

// importRoles applies the role lists of an older configuration to the
// users once, and records that it did: TASK_UI_ADMINS hold the admin
// flag and the rest lose it, TASK_UI_VIEWERS are viewers and the rest
// runners, and the TASK_UI_AGENT user is an MCP agent user whose
// mentions run the Notes agent. A list naming no user of storage
// changes nothing, so a typo can't take every admin away. With users
// yet to come, it waits for a start that has them.
func (a *App) importRoles(ctx context.Context) error {
	settings := storage.NewSettingStorage(a.DB)
	values, err := settings.Values(ctx)
	if err != nil || values[importedRoles] != "" {
		return err
	}
	if _, ok := os.LookupEnv("TASK_UI_AGENT_JOB"); ok {
		slog.Warn("TASK_UI_AGENT_JOB is not read: a note naming a user with Run instance Claude when mentioned runs the Notes agent as them")
	}
	lists := a.Config.Import
	users, err := a.Users.List(ctx)
	if err != nil {
		return err
	}
	if lists != nil && len(users) == 0 {
		return nil
	}
	if lists != nil {
		if err := applyRoles(ctx, a.Users, users, lists); err != nil {
			return err
		}
	}
	return settings.Set(ctx, map[string]string{importedRoles: "done"})
}

// applyRoles writes the role lists to the users, as importRoles says.
func applyRoles(ctx context.Context, store *storage.UserStorage, users []model.User, lists *RoleImport) error {
	admins := slices.ContainsFunc(users, func(u model.User) bool { return slices.Contains(lists.Admins, u.Name) })
	roles := admins || slices.ContainsFunc(users, func(u model.User) bool { return slices.Contains(lists.Viewers, u.Name) })
	for i := range users {
		user := &users[i]
		before := *user
		if admins {
			user.Admin = slices.Contains(lists.Admins, user.Name)
		}
		if roles {
			user.Role = model.RoleRunner
			if slices.Contains(lists.Viewers, user.Name) {
				user.Role = model.RoleViewer
			}
		}
		if lists.Agent != "" && user.Name == lists.Agent {
			user.Mcp, user.AgentReply = true, true
		}
		if *user == before {
			continue
		}
		if err := store.Update(ctx, user); err != nil {
			return err
		}
		slog.Info("user roles imported from the environment", "user", user.Name, "admin", user.Admin, "role", user.Role, "mcp", user.Mcp, "agent_reply", user.AgentReply)
	}
	slog.Info("role lists imported once: TASK_UI_ADMINS, TASK_UI_VIEWERS and TASK_UI_AGENT are not read again; remove them from the environment")
	return nil
}
