package api

import (
	"io"
	"net/http"

	"github.com/titpetric/task-ui/form"
)

// maxFormBody bounds the markdown one call reads, which is a note's reply.
const maxFormBody = 64 << 10

// formBlock is one form fence of the body: what it declares, or why the
// dialect refused it.
type formBlock struct {
	Form  *form.Form `json:"form,omitempty"`
	Error string     `json:"error,omitempty"`
}

// listForms answers with the form blocks of a markdown body, in order,
// each parsed or refused. The body is read and dropped: this reads no
// storage and writes none, and it exists so a client that holds a reply
// draws its forms without a yaml parser of its own. Takes prompt/view,
// the rule reading the reply takes.
func (h *Handlers) listForms(w http.ResponseWriter, r *http.Request) {
	if !h.require(w, r, "prompt/view") {
		return
	}
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxFormBody))
	if err != nil {
		fail(w, http.StatusRequestEntityTooLarge, "the body is longer than a reply")
		return
	}
	blocks := []formBlock{}
	for _, block := range form.Blocks(string(body)) {
		out := formBlock{Form: block.Form}
		if block.Err != nil {
			out.Error = block.Err.Error()
		}
		blocks = append(blocks, out)
	}
	writeJSON(w, http.StatusOK, map[string]any{"forms": blocks})
}
