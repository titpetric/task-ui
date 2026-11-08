package view_test

import (
	"testing"

	"github.com/titpetric/task-ui/frontend/view"
	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/model"
)

// The voice window calls an agent, which is a user whose runs answer as
// one: the Notes agent and MCP both on. A user a prompt may run as
// without them is no agent, and a page that offers no pick of user
// offers no agent.
func TestPage_Agents(t *testing.T) {
	t.Parallel()

	p := &view.Page{}
	assert.Len(t, p.Agents(), 0)

	p.PromptUsers = []model.User{
		{Name: "black", Role: model.RoleAdmin},
		{Name: "miguel", Role: model.RoleRunner, AgentReply: true, Mcp: true},
		{Name: "skyclaude", Role: model.RoleRunner, AgentReply: true},
		{Name: "consuela", Role: model.RoleRunner, AgentReply: true, Mcp: true},
	}
	names := []string{}
	for _, user := range p.Agents() {
		names = append(names, user.Name)
	}
	assert.Equal(t, []string{"miguel", "consuela"}, names)
}
