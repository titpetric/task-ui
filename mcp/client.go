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
)

// client calls the API in process, as the user on the context.
type client struct {
	api http.Handler
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
// an apiError with the API's message.
func (c *client) call(ctx context.Context, method, path string, body any) ([]byte, error) {
	var payload bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&payload).Encode(body); err != nil {
			return nil, err
		}
	}
	// The API is a router of its own; the route context of the request
	// that reached the server would steer it, so it starts without one.
	ctx = context.WithValue(ctx, chi.RouteCtxKey, (*chi.Context)(nil))
	req, err := http.NewRequestWithContext(ctx, method, path, &payload)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	c.api.ServeHTTP(rec, req)
	data := rec.Body.Bytes()
	if rec.Code >= http.StatusBadRequest {
		var e errorBody
		if json.Unmarshal(data, &e) != nil || e.Error == "" {
			e.Error = strings.TrimSpace(string(data))
		}
		return nil, &apiError{Status: rec.Code, Message: fmt.Sprintf("%s (%d)", e.Error, rec.Code), Body: data}
	}
	return data, nil
}

// get decodes the response to a GET into out.
func (c *client) get(ctx context.Context, path string, out any) error {
	data, err := c.call(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
