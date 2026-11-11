package phpfront_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/titpetric/task-ui/internal/assert"
	"github.com/titpetric/task-ui/phpfront"
)

func TestResult(t *testing.T) {
	t.Parallel()
	out := phpfront.Result{
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

	assert.Equal(t, "public", phpfront.PublicDir)
	assert.Equal(t, "index.php", phpfront.Index)
	assert.Equal(t, "modinfo.php", phpfront.ModInfo)
}
