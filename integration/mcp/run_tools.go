package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/titpetric/task-ui/api"
	"github.com/titpetric/task-ui/model"
)

// maxWait bounds how long a call waits for a run to end.
const maxWait = 15 * time.Minute

// pollEvery is how often a waiting call looks at the run.
var pollEvery = time.Second

type runArgs struct {
	Folder string `json:"folder"`
	Job    string `json:"job"`
	Force  bool   `json:"force"`
	Reason string `json:"reason"`
	Wait   int    `json:"wait"`
	Lines  *int   `json:"lines"`
}

// busyRun is the API's answer to a job that is running already.
type busyRun struct {
	Run api.RunView `json:"run"`
}

func runJob(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in runArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	data, err := c.call(ctx, http.MethodPost, "/runs", api.RunRequest{Folder: in.Folder, Job: in.Job, Force: in.Force, Reason: in.Reason, Cols: 120, Rows: 40})
	var e *apiError
	if errors.As(err, &e) && e.Status == http.StatusConflict {
		var busy busyRun
		if json.Unmarshal(e.Body, &busy) == nil && busy.Run.ID != "" {
			return "", fmt.Errorf("%s/%s is already running as run %s; force replaces it", in.Folder, in.Job, busy.Run.ID)
		}
	}
	if err != nil {
		return "", err
	}
	var view api.RunView
	if err := json.Unmarshal(data, &view); err != nil {
		return "", err
	}
	return showRun(ctx, c, view.ID, in.Wait, in.Lines)
}

type getRunArgs struct {
	ID    string `json:"id"`
	Lines *int   `json:"lines"`
}

func getRun(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in getRunArgs
	if err := decode(args, &in); err != nil {
		return "", err
	}
	if in.ID == "" {
		return "", errors.New("a run ID is needed")
	}
	return showRun(ctx, c, in.ID, 0, in.Lines)
}

// showRun waits up to wait seconds for a run to end, then describes it
// with the end of its output.
func showRun(ctx context.Context, c *client, id string, wait int, lines *int) (string, error) {
	path := "/runs/" + url.PathEscape(id)
	var view api.RunView
	if err := c.get(ctx, path, &view); err != nil {
		return "", err
	}
	if wait > 0 && !view.Run.Finished() {
		deadline := time.Now().Add(min(time.Duration(wait)*time.Second, maxWait))
		for !view.Run.Finished() && time.Now().Before(deadline) {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(pollEvery):
			}
			if err := c.get(ctx, path, &view); err != nil {
				return "", err
			}
		}
	}
	var b strings.Builder
	fmt.Fprintf(&b, "run %s: %s by %s, %s\n", view.ID, view.Path(), view.UserName, outcome(&view.Run))
	if view.Reason != "" {
		fmt.Fprintf(&b, "forced: %s\n", view.Reason)
	}
	if view.Title != "" {
		fmt.Fprintf(&b, "title: %s\n", view.Title)
	}
	n := 100
	if lines != nil {
		n = *lines
	}
	if n > 0 {
		data, err := c.call(ctx, http.MethodGet, path+"/frame.txt?all=1", nil)
		if err == nil {
			b.WriteString("\n")
			b.WriteString(tail(string(data), n))
		}
	}
	return b.String(), nil
}

type runsArgs struct {
	Folder string `json:"folder"`
	Job    string `json:"job"`
	State  string `json:"state"`
	Limit  int    `json:"limit"`
}

type runList struct {
	Runs []api.RunView `json:"runs"`
}

func listRuns(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	in := runsArgs{Limit: 20}
	if err := decode(args, &in); err != nil {
		return "", err
	}
	q := url.Values{}
	q.Set("folder", in.Folder)
	q.Set("job", in.Job)
	q.Set("state", in.State)
	q.Set("limit", fmt.Sprint(in.Limit))
	var out runList
	if err := c.get(ctx, "/runs?"+q.Encode(), &out); err != nil {
		return "", err
	}
	if len(out.Runs) == 0 {
		return "No runs.", nil
	}
	var b strings.Builder
	for _, view := range out.Runs {
		fmt.Fprintf(&b, "%s %s by %s: %s\n", view.ID, view.Path(), view.UserName, outcome(&view.Run))
		if view.Title != "" {
			fmt.Fprintf(&b, "  title: %s\n", view.Title)
		}
	}
	return b.String(), nil
}

// outcome says how a run went, when, and for how long.
func outcome(run *model.Run) string {
	when := ""
	if run.StartedAt != nil {
		when = " at " + run.StartedAt.Local().Format("2006-01-02 15:04:05")
	}
	took := run.Elapsed().Round(time.Second).String()
	switch {
	case run.State == model.StateQueued:
		return fmt.Sprintf("queued for %s, behind the run of its job before it, since%s", took, strings.TrimPrefix(when, " at"))
	case !run.Finished():
		return fmt.Sprintf("%s for %s, started%s", run.State, took, when)
	case run.Message != "":
		return fmt.Sprintf("%s (exit %d) in %s, started%s: %s", run.State, run.ExitCode, took, when, run.Message)
	}
	return fmt.Sprintf("%s (exit %d) in %s, started%s", run.State, run.ExitCode, took, when)
}

// tail returns the last n lines of text, without the blank ones at the end.
func tail(text string, n int) string {
	lines := strings.Split(strings.TrimRight(text, " \n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n") + "\n"
}
