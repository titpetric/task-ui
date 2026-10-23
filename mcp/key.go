package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/titpetric/task-ui/api"
)

// keyTool returns to an agent elsewhere an instance SSH key for the length
// of a task, its private half encrypted to an age recipient the agent
// made for the request, so the key never crosses in clear
// (docs/api.md).
func keyTool() Tool {
	return Tool{
		Name: "get_ssh_key",
		Description: "Get an instance SSH key to use for the length of a task: its private half, encrypted to the age " +
			"recipient you made for this request with age-keygen. Decrypt with age -d -i <identity>, keep it in /tmp, delete it after. " +
			"Takes key/read and MCP agent user; whoami lists the keys you may read.",
		InputSchema: object(map[string]Property{
			"name":      {Type: "string", Description: "The key's name, such as deploy_rsa4096."},
			"recipient": {Type: "string", Description: "An age X25519 recipient, age1..., for this request alone."},
		}, "name", "recipient"),
		call: getSSHKey,
	}
}

func getSSHKey(ctx context.Context, c *client, args json.RawMessage) (string, error) {
	var in struct {
		Name      string `json:"name"`
		Recipient string `json:"recipient"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", err
	}
	if in.Name == "" || in.Recipient == "" {
		return "", fmt.Errorf("name and recipient are needed")
	}
	var key api.KeyReadView
	if err := c.get(ctx, "/keys/"+url.PathEscape(in.Name)+"/private?recipient="+url.QueryEscape(in.Recipient), &key); err != nil {
		return "", err
	}
	kind := key.Algo
	if key.Bits > 0 {
		kind = fmt.Sprintf("%s %d", key.Algo, key.Bits)
	}
	return fmt.Sprintf("Key %s (%s, %s), encrypted to your recipient; %s.\n\n%s", key.Name, kind, key.Fingerprint, key.Note, key.Ciphertext), nil
}
