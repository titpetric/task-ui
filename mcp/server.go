// Package mcp serves task-ui to MCP clients, such as an agent that runs
// jobs: JSON-RPC over the streamable HTTP transport, at /api/mcp.
//
// The tools call the JSON API in process, as the user of the request, so
// what a client may see and run is what that user may in the browser:
// the same roles, allow lists and hours.
package mcp

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
)

// protocolVersions are the revisions of the protocol this server speaks,
// the newest first.
var protocolVersions = []string{"2025-06-18", "2025-03-26", "2024-11-05"}

// The JSON-RPC error codes the server answers with.
const (
	codeParse          = -32700
	codeInvalidRequest = -32600
	codeMethodNotFound = -32601
	codeInvalidParams  = -32602
)

// maxBody bounds a request body.
const maxBody = 1 << 20

// Server answers MCP requests over the API.
type Server struct {
	// API is the /api router the tools call.
	API http.Handler
	// Version is task-ui's, which the server reports.
	Version string
	tools   []Tool
}

// New returns a server whose tools call the API.
func New(api http.Handler, version string) *Server {
	return &Server{API: api, Version: version, tools: tools()}
}

// Tools returns the server's tools.
func (s *Server) Tools() []Tool {
	return s.tools
}

// request is a JSON-RPC request or notification; a notification has no ID.
type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// response is a JSON-RPC response: a result or an error.
type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *rpcError) Error() string {
	return e.Message
}

type initializeParams struct {
	ProtocolVersion string `json:"protocolVersion"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

type capabilities struct {
	Tools toolsCapability `json:"tools"`
}

type initializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    capabilities `json:"capabilities"`
	ServerInfo      serverInfo   `json:"serverInfo"`
	Instructions    string       `json:"instructions"`
}

type toolsListResult struct {
	Tools []Tool `json:"tools"`
}

type callParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

type content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type callResult struct {
	Content []content `json:"content"`
	IsError bool      `json:"isError"`
}

// empty is the result of ping.
type empty struct{}

// instructions tell a client what the server is for.
const instructions = "task-ui runs the jobs of folders (atkins, Taskfile, make, npm, scripts) and workflows, a chain of jobs, " +
	"as the user of the token. list_jobs and list_workflows show what there is, run_job and run_workflow start one " +
	"and can wait for the outcome, list_runs and get_run show how runs went and their output."

// ServeHTTP answers one JSON-RPC message per POST, in JSON. There is no
// stream to open with GET and no session to end with DELETE.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := s.serveHTTP(w, r); err != nil {
		rerr, ok := err.(*rpcError)
		if !ok {
			rerr = &rpcError{Code: codeParse, Message: err.Error()}
		}
		reply(w, response{Error: rerr})
	}
}

// serveHTTP answers a request, or returns why reading the message
// failed; the caller answers that without an ID.
func (s *Server) serveHTTP(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBody))
	if err != nil {
		return err
	}
	var req request
	if err := json.Unmarshal(data, &req); err != nil {
		return &rpcError{Code: codeParse, Message: "not a JSON-RPC message: " + err.Error()}
	}
	if req.JSONRPC != "2.0" || req.Method == "" {
		reply(w, response{ID: req.ID, Error: &rpcError{Code: codeInvalidRequest, Message: "not a JSON-RPC 2.0 request"}})
		return nil
	}
	if len(req.ID) == 0 {
		// A notification, such as notifications/initialized: nothing to say.
		w.WriteHeader(http.StatusAccepted)
		return nil
	}
	result, rerr := s.handle(r, req)
	reply(w, response{ID: req.ID, Result: result, Error: rerr})
	return nil
}

// handle answers a request by its method.
func (s *Server) handle(r *http.Request, req request) (any, *rpcError) {
	switch req.Method {
	case "initialize":
		var params initializeParams
		_ = json.Unmarshal(req.Params, &params)
		version := protocolVersions[0]
		if slices.Contains(protocolVersions, params.ProtocolVersion) {
			version = params.ProtocolVersion
		}
		return initializeResult{
			ProtocolVersion: version,
			ServerInfo:      serverInfo{Name: "task-ui", Version: s.Version},
			Instructions:    instructions,
		}, nil
	case "ping":
		return empty{}, nil
	case "tools/list":
		return toolsListResult{Tools: s.tools}, nil
	case "tools/call":
		var params callParams
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		tool := s.tool(params.Name)
		if tool == nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: "no tool " + params.Name}
		}
		text, err := tool.call(r.Context(), &client{api: s.API}, params.Arguments)
		if err != nil {
			return callResult{Content: []content{{Type: "text", Text: err.Error()}}, IsError: true}, nil
		}
		return callResult{Content: []content{{Type: "text", Text: text}}}, nil
	}
	return nil, &rpcError{Code: codeMethodNotFound, Message: "no method " + req.Method}
}

// tool returns the tool by name, or nil.
func (s *Server) tool(name string) *Tool {
	for i := range s.tools {
		if s.tools[i].Name == name {
			return &s.tools[i]
		}
	}
	return nil
}

// reply writes a JSON-RPC response.
func reply(w http.ResponseWriter, res response) {
	res.JSONRPC = "2.0"
	if res.ID == nil {
		res.ID = json.RawMessage("null")
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(res)
}
