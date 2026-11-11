package mcp

import (
	"encoding/json"

	"github.com/titpetric/task-ui/config"
)

// contextURI names `config/context.md` as a resource.
const contextURI = "taskui://context"

// resource is a document resources/read returns.
type resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

type resourcesListResult struct {
	Resources []resource `json:"resources"`
}

type readParams struct {
	URI string `json:"uri"`
}

type resourceText struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type readResult struct {
	Contents []resourceText `json:"contents"`
}

// resources are the server's documents: what task-ui is, each of its
// areas and the permission that governs it.
func resources() []resource {
	return []resource{{
		URI:         contextURI,
		Name:        "context.md",
		Description: "What task-ui is: each area, where it lives, and the permission that governs it.",
		MimeType:    "text/markdown",
	}}
}

// readResource answers resources/read.
func readResource(params json.RawMessage) (any, *rpcError) {
	var in readParams
	if err := json.Unmarshal(params, &in); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
	}
	if in.URI != contextURI {
		return nil, &rpcError{Code: codeInvalidParams, Message: "no resource " + in.URI}
	}
	return readResult{Contents: []resourceText{{URI: contextURI, MimeType: "text/markdown", Text: config.Context()}}}, nil
}
