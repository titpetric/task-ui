package phpscript_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpscript"
)

func TestResult(t *testing.T) {
	t.Parallel()
	out := phpscript.Result{
		Body:   []byte("<p>a thread</p>"),
		Status: http.StatusOK,
		Header: http.Header{"Content-Type": []string{"text/html"}},
		Took:   3 * time.Millisecond,
		File:   false,
	}
	assert.Equal(t, "<p>a thread</p>", string(out.Body))
	assert.Equal(t, http.StatusOK, out.Status)
	assert.Equal(t, "text/html", out.Header.Get("Content-Type"))
	assert.Equal(t, 3*time.Millisecond, out.Took)
	assert.False(t, out.File)

	assert.Equal(t, "public", phpscript.Public)
	assert.Equal(t, "index.php", phpscript.Index)
	assert.Equal(t, "modinfo.php", phpscript.ModInfo)
}
