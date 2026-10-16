package server

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/titpetric/task-ui/server/model"
)

// sessionCookie is the name of the session cookie.
const sessionCookie = "task_ui_session"

var errInvalidToken = errors.New("invalid session")

// sessions signs and verifies the session cookie: an HS256 JWT holding the
// username. The role is looked up from the configuration on every request,
// so a user removed from task-ui.yml loses access on their next click.
type sessions struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func newSessions(secret string, ttl time.Duration) *sessions {
	key := []byte(secret)
	if len(key) == 0 {
		key = make([]byte, 32)
		_, _ = rand.Read(key)
	}
	return &sessions{secret: key, ttl: ttl, now: time.Now}
}

func segment(v any) string {
	data, _ := json.Marshal(v)
	return base64.RawURLEncoding.EncodeToString(data)
}

func (s *sessions) sign(input string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Issue returns a session token for a user.
func (s *sessions) Issue(username string) string {
	now := s.now()
	signed := segment(map[string]string{"alg": "HS256", "typ": "JWT"}) + "." +
		segment(map[string]any{"sub": username, "iat": now.Unix(), "exp": now.Add(s.ttl).Unix()})
	return signed + "." + s.sign(signed)
}

// Verify returns the username a token was issued to.
func (s *sessions) Verify(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 || !hmac.Equal([]byte(s.sign(parts[0]+"."+parts[1])), []byte(parts[2])) {
		return "", errInvalidToken
	}
	data, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errInvalidToken
	}
	var claims struct {
		Sub string `json:"sub"`
		Exp int64  `json:"exp"`
	}
	if err := json.Unmarshal(data, &claims); err != nil || claims.Sub == "" {
		return "", errInvalidToken
	}
	if claims.Exp != 0 && s.now().Unix() >= claims.Exp {
		return "", errInvalidToken
	}
	return claims.Sub, nil
}

type userKey struct{}

// currentUser returns the user of a request, nil for nobody.
func currentUser(r *http.Request) *model.User {
	user, _ := r.Context().Value(userKey{}).(*model.User)
	return user
}

// loginRequired reports whether requests carry a session. Login is the
// only way in, unless the development auto login is set.
func (svc *Server) loginRequired() bool {
	return svc.config.AutoLogin == ""
}

// hasUsers reports whether task-ui.yml names anyone.
func (svc *Server) hasUsers() bool {
	return len(svc.file.Get().Users) > 0
}

// lookupUser returns the configured user by name.
func (svc *Server) lookupUser(name string) *model.User {
	for _, user := range svc.file.Get().Users {
		if user.Name == name {
			u := user
			return &u
		}
	}
	return nil
}

// authenticate resolves the user of a request from its session cookie, or
// from the auto login when one is configured.
func (svc *Server) authenticate(r *http.Request) *model.User {
	if !svc.loginRequired() {
		return svc.lookupUser(svc.config.AutoLogin)
	}
	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		return nil
	}
	name, err := svc.sessions.Verify(cookie.Value)
	if err != nil {
		return nil
	}
	return svc.lookupUser(name)
}

// requireUser sends visitors without a session to the login screen, and
// answers API and websocket requests with 401.
func (svc *Server) requireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := svc.authenticate(r)
		if user == nil {
			if r.Method == http.MethodGet && !strings.HasPrefix(r.URL.Path, "/api/") && !strings.HasPrefix(r.URL.Path, "/ws/") && svc.loginRequired() {
				http.Redirect(w, r, "/login?back="+url.QueryEscape(r.URL.RequestURI()), http.StatusSeeOther)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

// sameOrigin rejects state changing requests from other sites. The session
// cookie is SameSite=Strict as well; this covers browsers that ignore it.
func sameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead && !originMatches(r) {
			http.Error(w, "cross origin request refused", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func originMatches(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Non-browser clients and older browsers on same origin
		// form posts send no Origin; fall back to the Referer.
		origin = r.Header.Get("Referer")
		if origin == "" {
			return true
		}
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return sameHost(u.Host, r)
}

// sameHost reports whether host is the host the request was made to. A
// reverse proxy that rewrites Host passes the original in X-Forwarded-Host.
func sameHost(host string, r *http.Request) bool {
	if host == "" {
		return false
	}
	if host == r.Host {
		return true
	}
	for _, forwarded := range strings.Split(r.Header.Get("X-Forwarded-Host"), ",") {
		if strings.TrimSpace(forwarded) == host {
			return true
		}
	}
	return false
}

// safeBack returns a local redirect target, or /.
func safeBack(back string) string {
	if back == "" || !strings.HasPrefix(back, "/") || strings.HasPrefix(back, "//") || strings.HasPrefix(back, "/\\") {
		return "/"
	}
	return back
}

// Login serves and handles the sign in form.
func (svc *Server) Login(w http.ResponseWriter, r *http.Request) {
	data := svc.pageData(r, "Sign in")
	data.Back = safeBack(r.FormValue("back"))

	if !svc.loginRequired() {
		http.Redirect(w, r, data.Back, http.StatusSeeOther)
		return
	}
	if !svc.hasUsers() {
		data.Error = "No users are configured. Add one to task-ui.yml, with a password from `task-ui hash`, and sign in."
		w.WriteHeader(http.StatusServiceUnavailable)
		svc.render(w, r, "login.tpl", data)
		return
	}

	if r.Method == http.MethodPost {
		username := strings.TrimSpace(r.PostFormValue("username"))
		password := r.PostFormValue("password")
		user := svc.lookupUser(username)
		if user != nil && bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) == nil {
			http.SetCookie(w, &http.Cookie{
				Name:     sessionCookie,
				Value:    svc.sessions.Issue(user.Name),
				Path:     "/",
				HttpOnly: true,
				Secure:   isHTTPS(r),
				SameSite: http.SameSiteStrictMode,
				MaxAge:   int(svc.config.SessionTTL.Seconds()),
			})
			http.Redirect(w, r, data.Back, http.StatusSeeOther)
			return
		}
		if user == nil {
			// Spend the same time as a real comparison, so a missing
			// user is not distinguishable by timing.
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		}
		data.Username = username
		data.Error = "Unknown user or wrong password."
		w.WriteHeader(http.StatusUnauthorized)
	}

	svc.render(w, r, "login.tpl", data)
}

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("task-ui"), bcrypt.DefaultCost)

// Logout clears the session.
func (svc *Server) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}
