package runner

import (
	"context"
	"os"
	"os/exec"
	"strconv"

	"github.com/titpetric/task-ui/model"
)

// command builds the process: bash -l for the login environment, the
// job's argv passed as arguments so the shell never parses a name, and
// the request's args after --, the same way; the share of the server's
// environment the parent environment passes, the secrets and the run
// facts in the environment, HOME at the workspace's home project, the
// artifact directory ready. The process starts in the request's
// directory, or the folder's.
func (m *Manager) command(ctx context.Context, session *Session) (*exec.Cmd, error) {
	f, job, run := session.Folder, session.Job, session.Run
	args := append([]string{"-l", "-c", `exec "$0" "$@"`}, job.Command...)
	// A host job's command carries its args already, as the remote
	// command of ssh.
	if len(session.args) > 0 && !job.IsHost() {
		args = append(append(args, "--"), session.args...)
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = f.Dir
	if session.dir != "" {
		cmd.Dir = session.dir
	}
	cmd.Env = m.parentEnv(ctx)
	cmd.Env = append(cmd.Env, HomeEnv(cmd.Env, session.home)...)
	// Every run gets its folder's secrets, Claude's built-ins too, and
	// the log of each secret records the run.
	if m.opts.Secrets != nil {
		secrets, err := m.opts.Secrets.Reaching(ctx, f.ID)
		if err != nil {
			return nil, err
		}
		reads := make([]model.SecretRead, 0, len(secrets))
		for _, secret := range secrets {
			cmd.Env = append(cmd.Env, secret.Name+"="+secret.Content)
			reads = append(reads, model.SecretRead{SecretID: secret.ID, Name: secret.Name, RunID: run.ID, Folder: run.Folder, Dir: run.Dir, Job: run.Job, UserName: run.UserName})
		}
		if m.opts.SecretReads != nil {
			if err := m.opts.SecretReads.Record(ctx, reads); err != nil {
				return nil, err
			}
		}
	}
	if m.opts.Integrations != nil {
		vars, err := m.opts.Integrations.Env(ctx, run.UserName, session.role)
		if err != nil {
			return nil, err
		}
		for _, key := range model.SortedKeys(vars) {
			cmd.Env = append(cmd.Env, key+"="+vars[key])
		}
	}
	cmd.Env = append(cmd.Env, session.env...)
	cmd.Env = append(cmd.Env,
		"TERM=xterm-256color",
		"COLORTERM=truecolor",
		"COLUMNS="+strconv.Itoa(int(run.TermCols)),
		"LINES="+strconv.Itoa(int(run.TermRows)),
		"TASK_UI=1",
		"TASK_UI_USER="+run.UserName,
		"TASK_UI_ROLE="+session.role,
		"TASK_UI_FOLDER="+f.ID,
		"TASK_UI_JOB="+job.ID,
		"TASK_UI_RUN_ID="+run.ID,
		"TASK_UI_HOST_PWD="+m.hostPath(cmd.Dir),
	)
	if m.opts.HostData != "" {
		cmd.Env = append(cmd.Env, "TASK_UI_HOST_DATA="+m.opts.HostData)
	}
	if session.permissions != "" {
		cmd.Env = append(cmd.Env, "TASK_UI_PERMISSIONS="+session.permissions)
	}
	if session.workspace != "" {
		cmd.Env = append(cmd.Env,
			"TASK_UI_WORKSPACE="+session.workspaceName,
			"TASK_UI_WORKSPACE_DIR="+session.workspace,
		)
	}
	if dir := m.artifactDir(run); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, "TASK_UI_ARTIFACTS="+dir)
	}
	return cmd, nil
}
