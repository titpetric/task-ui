package phpscript

import (
	"net/http"
	"time"
)

// Result is what one render produced. The body reaches the response unwritten,
// because a module's output is the body of a task-ui page and the layout goes
// around it; a handler that serves a file writes Body as it stands.
type Result struct {
	// Body is what the script echoed, or the bytes of a file under public/.
	Body []byte

	// Status is what the script staged, 200 when it staged none.
	Status int

	// Header holds the headers the script staged. Content-Type is set for a
	// file served from public/ and left to the layout for a rendered page.
	Header http.Header

	// Took is how long the render ran, which the page footer prints.
	Took time.Duration

	// File reports a path under the module's public/ tree, served as bytes
	// with no layout around it.
	File bool
}
