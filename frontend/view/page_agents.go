package view

import (
	"github.com/titpetric/task-ui/model"
)

// Agents returns the users the voice window may call, by name: the ones
// whose runs answer as an agent, out of the users a prompt may run as.
// The window draws no pick where there is none.
func (p *Page) Agents() []model.User {
	var out []model.User
	for _, user := range p.PromptUsers {
		if user.AgentReply && user.Mcp && user.Name != "" {
			out = append(out, user)
		}
	}
	return out
}
