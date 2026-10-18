package api_test

import (
	"bufio"
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
)

// sse opens an event stream as a user and returns the response.
func (f *fixture) sse(t *testing.T, ctx context.Context, user, topic string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.server.URL+"/sse/"+topic, nil)
	assert.NoError(t, err)
	req.Header.Set("X-User", user)
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	return res
}

// A workflow's topic streams its run: the run, each step as it goes,
// and the end, as server-sent events.
func TestSSE_Workflow(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	id := f.createWorkflow(t, asOps, "chain", chain("infra", "plan", "infra/edge", "deploy"))

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res := f.sse(t, ctx, asViewer, "workflow:"+id) // watching needs seeing only
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, res.Header.Get("Content-Type"), "text/event-stream")

	started := f.call(t, asOps, http.MethodPost, "/workflows/"+id+"/runs", nil)
	assert.Equal(t, http.StatusCreated, started.Status, string(started.Raw))

	names, datas := []string{}, []string{}
	scanner := bufio.NewScanner(res.Body)
	name := ""
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "event: "):
			name = strings.TrimPrefix(line, "event: ")
		case strings.HasPrefix(line, "data: "):
			names = append(names, name)
			datas = append(datas, strings.TrimPrefix(line, "data: "))
		}
		if name == "done" && len(datas) == len(names) && len(names) > 0 && names[len(names)-1] == "done" {
			break
		}
	}
	assert.Equal(t, "run", names[0])
	assert.Contains(t, datas[0], `"run_id":"`+started.str("id")+`"`)
	assert.Equal(t, "done", names[len(names)-1])
	assert.Contains(t, datas[len(datas)-1], `"state":"ok"`)
	steps := 0
	for i, n := range names {
		if n == "step" {
			steps++
			assert.Contains(t, datas[i], `"position":`)
			assert.Contains(t, datas[i], `"state":"`)
		}
	}
	assert.True(t, steps >= 4, "each step reports running and ok: %v", names)
	assert.Contains(t, strings.Join(datas, "\n"), `"state":"running"`)
	assert.Contains(t, strings.Join(datas, "\n"), `"folder":"infra/edge"`)
}

// The runs topic streams every run as it starts and ends, of the
// folders the user sees only.
func TestSSE_Runs(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	res := f.sse(t, ctx, asLimited, "runs")
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	f.finished(t, f.start(t, asAdmin, "private", "deploy")) // limited does not see private
	id := f.start(t, asOps, "infra", "plan")

	scanner := bufio.NewScanner(res.Body)
	datas := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			datas = append(datas, strings.TrimPrefix(line, "data: "))
			if strings.Contains(line, `"state":"ok"`) {
				break
			}
		}
	}
	all := strings.Join(datas, "\n")
	assert.Contains(t, datas[0], `"id":"`+id+`"`)
	assert.Contains(t, datas[0], `"state":"running"`)
	assert.Contains(t, all, `"folder":"infra"`)
	assert.NotContains(t, all, `"folder":"private"`)
}

// A topic that is not a workflow the user sees answers 404.
func TestSSE_Hidden(t *testing.T) {
	t.Parallel()
	f := newFixture(t)
	private := f.createWorkflow(t, asAdmin, "private", chain("private", "deploy"))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, topic := range []string{"workflow:" + private, "workflow:nope", "nope:" + private, "workflow"} {
		res := f.sse(t, ctx, asLimited, topic)
		res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode, topic)
	}
}
