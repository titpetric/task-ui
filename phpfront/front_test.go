package phpfront

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	chi "github.com/go-chi/chi/v5"
)

// TestMountStripsThePrefix: the PHP declares paths from the root, the router
// is mounted under a prefix, and the page answers.
func TestMountStripsThePrefix(t *testing.T) {
	front := testFront(t, "/development")
	router := chi.NewRouter()
	router.Route("/development", func(r chi.Router) {
		if err := front.Mount(context.Background(), r); err != nil {
			t.Fatalf("mounting: %v", err)
		}
	})

	for _, tc := range []struct {
		name   string
		method string
		path   string
		status int
		want   string
	}{
		{"a page", http.MethodGet, "/development/probe", http.StatusOK, "probe served"},
		{"the prefix reaches the base", http.MethodGet, "/development/probe", http.StatusOK, "base /development"},
		{"a path parameter", http.MethodGet, "/development/probe/01HZ", http.StatusOK, "id=01HZ"},
		{"a post with a parameter", http.MethodPost, "/development/probe/01HZ/comments", http.StatusOK, "id=01HZ"},
		{"a methodless route takes get", http.MethodGet, "/development/probe/both", http.StatusOK, "GET"},
		{"a methodless route takes post", http.MethodPost, "/development/probe/both", http.StatusOK, "POST"},
		{"the catch all answers the rest", http.MethodGet, "/development/nope", http.StatusNotFound, "nothing at"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
			if w.Code != tc.status {
				t.Errorf("status: want %d, got %d: %s", tc.status, w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tc.want) {
				t.Errorf("body: want %q in %q", tc.want, w.Body.String())
			}
		})
	}
}

// TestRequestURIKeepsThePrefix: the request is not rewritten, so $_SERVER holds
// the prefix. Only Task\Route parses it, against Task::base().
func TestRequestURIKeepsThePrefix(t *testing.T) {
	front := testFront(t, "/development")
	router := chi.NewRouter()
	router.Route("/development", func(r chi.Router) {
		if err := front.Mount(context.Background(), r); err != nil {
			t.Fatalf("mounting: %v", err)
		}
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/development/probe/01HZ", nil))
	if want := "/development/probe/01HZ"; !strings.Contains(w.Body.String(), want) {
		t.Errorf("want %q in %q", want, w.Body.String())
	}
}

// TestDeclarationSurvivesTheNextRequest: a pooled runtime is reset between
// requests, so a mod_ file may declare its class and dispatch in one file.
func TestDeclarationSurvivesTheNextRequest(t *testing.T) {
	front := testFront(t, "")
	handler, err := front.Routes(context.Background())
	if err != nil {
		t.Fatalf("routes: %v", err)
	}

	for i := range 3 {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/probe", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d: %s", i, w.Code, w.Body.String())
		}
		if want := "probe served 1"; !strings.Contains(w.Body.String(), want) {
			t.Errorf("request %d: want %q in %q", i, want, w.Body.String())
		}
	}
}

// TestMountWithoutPrefix is the state after the cutover: the same tree, no prefix.
func TestMountWithoutPrefix(t *testing.T) {
	front := testFront(t, "")
	handler, err := front.Routes(context.Background())
	if err != nil {
		t.Fatalf("routes: %v", err)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/probe", nil))
	if want := "base "; !strings.HasSuffix(w.Body.String(), want) {
		t.Errorf("want the body to end in an empty base, got %q", w.Body.String())
	}
}

// A link built in Go carries the prefix, so the two sides agree on one path.
func TestFront_Href(t *testing.T) {
	front := testFront(t, "/development")
	if want, got := "/development/settings", front.Href("/settings"); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := "/development/settings", front.Href("settings"); want != got {
		t.Errorf("want %q, got %q", want, got)
	}
	if want, got := "/settings", testFront(t, "").Href("/settings"); want != got {
		t.Errorf("after the cutover: want %q, got %q", want, got)
	}
	var none *Front
	if want, got := "/settings", none.Href("/settings"); want != got {
		t.Errorf("nil front: want %q, got %q", want, got)
	}
}

// The static files answer under their own prefix, which no page route claims.
func TestFront_Assets(t *testing.T) {
	front := testFront(t, "/development")
	if front.Assets() == nil {
		t.Fatal("no static tree")
	}
	if len(front.Assets().Version()) != 12 {
		t.Errorf("version: %q", front.Assets().Version())
	}
	var none *Front
	if none.Assets() != nil {
		t.Error("a nil front has no static tree")
	}
}

// testFront returns the front end over the routing fixture.
func testFront(t *testing.T, base string) *Front {
	t.Helper()
	opts := NewOptions("testdata/routes")
	opts.Base = base
	opts.OnError = func(err error) { t.Errorf("router: %v", err) }
	front, err := New(opts)
	if err != nil {
		t.Fatalf("new: %v", err)
	}
	return front
}
