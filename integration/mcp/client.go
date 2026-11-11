package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"github.com/titpetric/task-ui/access"
)

// client calls the API in process, as the user on the context.
type client struct {
	api http.Handler
	// base is the public base link of the instance, for the links a tool
	// returns.
	base string
}

// apiError is an API answer that is not a success.
type apiError struct {
	Status  int
	Message string
	Body    []byte
}

func (e *apiError) Error() string {
	return e.Message
}

// errorBody is how the API says what went wrong.
type errorBody struct {
	Error string `json:"error"`
}

// call sends a request to the API and returns the body of a success, or
// an apiError with the API's message; a refusal, which names the
// permission it needs, points to whoami.
func (c *client) call(ctx context.Context, method, path string, body any) ([]byte, error) {
	var payload bytes.Buffer
	contentType := ""
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return nil, err
		}
		contentType = "application/json"
	}
	data, _, err := c.send(ctx, method, path, contentType, payload.Bytes())
	return data, err
}

// send sends a request with a body of the type given, none when empty,
// and returns the body and the headers of a success, or an apiError.
func (c *client) send(ctx context.Context, method, path, contentType string, body []byte) ([]byte, http.Header, error) {
	// The API is a router of its own; the route context of the request
	// that reached the server would steer it, so it starts without one.
	ctx = context.WithValue(ctx, chi.RouteCtxKey, (*chi.Context)(nil))
	ctx = access.WithVia(ctx, access.ViaMCP)
	req, err := http.NewRequestWithContext(ctx, method, path, bytes.NewReader(body))
	if err != nil {
		return nil, nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	c.api.ServeHTTP(rec, req)
	data := rec.Body.Bytes()
	if rec.Code >= http.StatusBadRequest {
		var e errorBody
		if json.Unmarshal(data, &e) != nil || e.Error == "" {
			e.Error = strings.TrimSpace(string(data))
		}
		message := fmt.Sprintf("%s (%d)", e.Error, rec.Code)
		if rec.Code == http.StatusForbidden {
			message += "; call whoami to see what you can do"
		}
		return nil, nil, &apiError{Status: rec.Code, Message: message, Body: data}
	}
	return data, rec.Header(), nil
}

// get decodes the response to a GET into out.
func (c *client) get(ctx context.Context, path string, out any) error {
	data, err := c.call(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
