package app_test

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/titpetric/task-ui/app"
	"github.com/titpetric/task-ui/internal/assert"
)

// migrations is every migration of the instance, in the order a run applies
// them, which is the order and the spelling every existing instance has
// already recorded in its migrations table.
//
// This list is the one thing standing between a refactor and a broken
// production database. A file's recorded name is its name here: move a
// migration between owners and nothing changes, but rename its file, or let
// two owners end up with one base name, and an instance that already ran it
// runs it again. The second CREATE TABLE is conditional and survives; the
// second ALTER TABLE of a column that is already there does not.
//
// So a change to this list is a deliberate edit of 86 strings, which a review
// sees, rather than a number that quietly moved.
var migrations = []string{
	"artifact.up.sql",
	"artifact_commit.up.sql",
	"artifact_user.up.sql",
	"env_file.up.sql",
	"key_info.up.sql",
	"key_info_backup.up.sql",
	"key_read.up.sql",
	"key_role.up.sql",
	"key_user.up.sql",
	"module.up.sql",
	"note.up.sql",
	"note_about_run.up.sql",
	"note_about_workflow_run.up.sql",
	"note_area.up.sql",
	"note_edited.up.sql",
	"note_workspace.up.sql",
	"permission_policy.up.sql",
	"permission_policy_runner_artifact.up.sql",
	"pin.up.sql",
	"project.up.sql",
	"project_hours_drop.up.sql",
	"project_root_only.up.sql",
	"project_root_only_shallow.up.sql",
	"project_sidebar.up.sql",
	"refinement_artifact.up.sql",
	"refinement_comment.up.sql",
	"refinement_correction.up.sql",
	"refinement_thread.up.sql",
	"run.up.sql",
	"run_dir_args.up.sql",
	"run_summary.up.sql",
	"run_target.up.sql",
	"run_target_host_job.up.sql",
	"run_target_queued.up.sql",
	"run_title.up.sql",
	"schedule.up.sql",
	"schedule_target.up.sql",
	"schedule_target_host_job.up.sql",
	"schedule_triage.up.sql",
	"schedule_utc.up.sql",
	"schedule_workflow.up.sql",
	"secret.up.sql",
	"secret_is_secret.up.sql",
	"secret_read.up.sql",
	"setting.up.sql",
	"setting_parent_env.up.sql",
	"sidebar.up.sql",
	"ssh_key_policy.up.sql",
	"ssh_key_user.up.sql",
	"tool.up.sql",
	"user.up.sql",
	"user_auth_password.up.sql",
	"user_auth_token.up.sql",
	"user_auth_token_expiry.up.sql",
	"user_auth_token_kind.up.sql",
	"user_email.up.sql",
	"user_flags.up.sql",
	"user_session.up.sql",
	"workflow.up.sql",
	"workflow_description.up.sql",
	"workflow_prompt.up.sql",
	"workflow_run.up.sql",
	"workflow_run_answer.up.sql",
	"workflow_run_as.up.sql",
	"workflow_run_step.up.sql",
	"workflow_runs_over.up.sql",
	"workflow_sidebar.up.sql",
	"workflow_step.up.sql",
	"workflow_step_enabled.up.sql",
	"workflow_step_prompt.up.sql",
	"workflow_step_ref.up.sql",
	"workflow_step_target.up.sql",
	"workflow_step_target_host_job.up.sql",
	"workflow_step_with_kind.up.sql",
	"workflow_step_with_kind_check.up.sql",
	"workflow_step_workspace_none.up.sql",
	"workspace.up.sql",
	"workspace_home_project.up.sql",
	"workspace_home_url.up.sql",
	"workspace_hours.up.sql",
	"workspace_key_grant.up.sql",
	"workspace_shallow.up.sql",
	"workspace_sidebar_limit.up.sql",
	"workspace_sort.up.sql",
	"workspaces_default_internal.up.sql",
	"workspaces_internal_title.up.sql",
}

// The merged set is exactly what the owners carry, flat and in order. It also
// proves the merge does not panic, which is how a base name carried by two
// owners shows up: app.Migrations is called at every start and by every
// maintenance command, so this test is the one that fails first.
func TestMigrations(t *testing.T) {
	t.Parallel()
	files, err := fs.Glob(app.Migrations(), "*.up.sql")
	assert.NoError(t, err)
	assert.Equal(t, migrations, files)
}

// Nothing but migrations reaches the merged set: a module's schema/ holds the
// generated schema.puml and schema.svg beside the SQL, and a run that handed
// those to mig would try to execute them.
func TestMigrationsCarryNothingElse(t *testing.T) {
	t.Parallel()
	entries, err := fs.ReadDir(app.Migrations(), ".")
	assert.NoError(t, err)
	assert.Len(t, entries, len(migrations))
	for _, entry := range entries {
		assert.True(t, strings.HasSuffix(entry.Name(), ".up.sql"), entry.Name())
	}
}

// Every owner is represented, which is what catches a //go:embed that stopped
// matching: the merge is one filesystem and a missing owner is a missing table
// rather than a missing file.
func TestMigrationsCoverEveryOwner(t *testing.T) {
	t.Parallel()
	for _, name := range []string{
		"note.up.sql",     // phpfront, a module's pages
		"key_info.up.sql", // keystore
		"user_session.up.sql",
		"module.up.sql", // phpscript, the module registry
	} {
		data, err := fs.ReadFile(app.Migrations(), name)
		assert.NoError(t, err, name)
		assert.Contains(t, string(data), "CREATE TABLE IF NOT EXISTS", name)
	}
}

// The two columns an ALTER added while the phpscript module owned the
// refinement tables are declared in the CREATE TABLE. A database that already
// holds the table fails that ALTER a second time, and SQLite has no ADD COLUMN
// IF NOT EXISTS to soften it, so the column belongs where every statement of
// these four files is conditional.
func TestMigrationsRefinementColumnsAreDeclared(t *testing.T) {
	t.Parallel()
	data, err := fs.ReadFile(app.Migrations(), "refinement_comment.up.sql")
	assert.NoError(t, err)
	assert.Contains(t, string(data), "kind TEXT NOT NULL DEFAULT ''")
	assert.Contains(t, string(data), "collapsed_into CHAR(26) NOT NULL DEFAULT ''")
}
