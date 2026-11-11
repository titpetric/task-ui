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

// The sign in page answers before a sign in and every other page does not,
// which is the boundary the mount draws: a bug here would hand the tree to
// anyone, so the two sides of it are checked rather than reviewed.
func TestFront_Guarded(t *testing.T) {
	for _, base := range []string{"/development", ""} {
		front := testFront(t, base)
		gate := &testGate{}
		drawn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("the page"))
		})
		handler := front.Guarded(gate, drawn)

		for _, tc := range []struct {
			name   string
			path   string
			status int
			want   string
		}{
			{"the sign in page is named, not required", base + "/login", http.StatusOK, "attached: the page"},
			{"a query does not open a path", base + "/login?back=%2Fruns", http.StatusOK, "attached: the page"},
			{"every other page is required", base + "/notes", http.StatusUnauthorized, "required"},
			{"the root is required", base + "/", http.StatusUnauthorized, "required"},
			{"a path that ends in the public one is not it", base + "/users/login", http.StatusUnauthorized, "required"},
		} {
			t.Run(base+" "+tc.name, func(t *testing.T) {
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
				if w.Code != tc.status {
					t.Errorf("status: want %d, got %d", tc.status, w.Code)
				}
				if !strings.Contains(w.Body.String(), tc.want) {
					t.Errorf("want %q in %q", tc.want, w.Body.String())
				}
			})
		}
	}

	// A nil front mounts nothing, so it guards nothing either.
	var none *Front
	drawn := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("the page"))
	})
	w := httptest.NewRecorder()
	none.Guarded(&testGate{}, drawn).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notes", nil))
	if want := "the page"; w.Body.String() != want {
		t.Errorf("nil front: want %q, got %q", want, w.Body.String())
	}
}

// testGate stands in for the Guard: Require refuses, Attach says it named the
// reader and passes the request on.
type testGate struct{}

var _ Gate = (*testGate)(nil)

func (g *testGate) Require(http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "required", http.StatusUnauthorized)
	})
}

func (g *testGate) Attach(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("attached: "))
		next.ServeHTTP(w, r)
	})
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
