package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

const testAUD = "test-audience-tag"

type accessFixture struct {
	srv    *httptest.Server
	key    *rsa.PrivateKey
	router *gin.Engine
}

func newAccessFixture(t *testing.T, allowed ...string) *accessFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "k1", Algorithm: "RS256", Use: "sig"}}}

	mux := http.NewServeMux()
	mux.HandleFunc("/cdn-cgi/access/certs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	mw, err := NewAccessAuth(context.Background(), AccessConfig{
		TeamDomain:    srv.URL,
		Audience:      testAUD,
		AllowedEmails: allowed,
	})
	if err != nil {
		t.Fatal(err)
	}
	r := gin.New()
	r.GET("/admin", mw, func(c *gin.Context) { c.String(http.StatusOK, AdminEmail(c)) })
	return &accessFixture{srv: srv, key: key, router: r}
}

type tokenOpts struct {
	issuer   string
	aud      string
	email    string
	expires  time.Time
	signWith *rsa.PrivateKey
	kid      string
}

func (f *accessFixture) token(t *testing.T, o tokenOpts) string {
	t.Helper()
	if o.issuer == "" {
		o.issuer = f.srv.URL
	}
	if o.aud == "" {
		o.aud = testAUD
	}
	if o.expires.IsZero() {
		o.expires = time.Now().Add(time.Hour)
	}
	if o.signWith == nil {
		o.signWith = f.key
	}
	if o.kid == "" {
		o.kid = "k1"
	}
	signer, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.RS256, Key: jose.JSONWebKey{Key: o.signWith, KeyID: o.kid}},
		(&jose.SignerOptions{}).WithType("JWT"),
	)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := jwt.Signed(signer).
		Claims(jwt.Claims{Issuer: o.issuer, Audience: jwt.Audience{o.aud}, Expiry: jwt.NewNumericDate(o.expires), IssuedAt: jwt.NewNumericDate(time.Now())}).
		Claims(map[string]any{"email": o.email}).
		Serialize()
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func (f *accessFixture) do(token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	if token != "" {
		req.Header.Set(accessJWTHeader, token)
	}
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	return w
}

func TestAccessAuth(t *testing.T) {
	f := newAccessFixture(t, "Dillon@Example.com")
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name  string
		token string
		want  int
		body  string
	}{
		{"valid token", f.token(t, tokenOpts{email: "dillon@example.com"}), http.StatusOK, "dillon@example.com"},
		{"email match is case-insensitive", f.token(t, tokenOpts{email: "DILLON@EXAMPLE.COM"}), http.StatusOK, "dillon@example.com"},
		{"missing header", "", http.StatusUnauthorized, ""},
		{"garbage token", "not.a.jwt", http.StatusUnauthorized, ""},
		{"wrong audience", f.token(t, tokenOpts{email: "dillon@example.com", aud: "someone-elses-app"}), http.StatusUnauthorized, ""},
		{"wrong issuer", f.token(t, tokenOpts{email: "dillon@example.com", issuer: "https://evil.cloudflareaccess.com"}), http.StatusUnauthorized, ""},
		{"expired", f.token(t, tokenOpts{email: "dillon@example.com", expires: time.Now().Add(-time.Hour)}), http.StatusUnauthorized, ""},
		{"signed by an unknown key", f.token(t, tokenOpts{email: "dillon@example.com", signWith: otherKey}), http.StatusUnauthorized, ""},
		{"valid token, email not on allow-list", f.token(t, tokenOpts{email: "attacker@example.com"}), http.StatusForbidden, ""},
		{"valid token, no email claim", f.token(t, tokenOpts{email: ""}), http.StatusForbidden, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := f.do(tc.token)
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d (body %q)", w.Code, tc.want, w.Body.String())
			}
			if tc.body != "" && w.Body.String() != tc.body {
				t.Errorf("body = %q, want %q", w.Body.String(), tc.body)
			}
		})
	}
}

func TestAccessAuthRejectsUnsignedAndHMACTokens(t *testing.T) {
	f := newAccessFixture(t, "dillon@example.com")

	// alg=none, hand-built.
	none := "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
		"eyJpc3MiOiJ4IiwiYXVkIjoidGVzdC1hdWRpZW5jZS10YWciLCJlbWFpbCI6ImRpbGxvbkBleGFtcGxlLmNvbSIsImV4cCI6OTk5OTk5OTk5OX0."
	if w := f.do(none); w.Code != http.StatusUnauthorized {
		t.Errorf("alg=none token: status = %d, want 401", w.Code)
	}

	// HS256 token "signed" with a made-up secret (algorithm-confusion attempt).
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte("0123456789abcdef0123456789abcdef")}, nil)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := jwt.Signed(signer).Claims(jwt.Claims{Issuer: f.srv.URL, Audience: jwt.Audience{testAUD}, Expiry: jwt.NewNumericDate(time.Now().Add(time.Hour))}).
		Claims(map[string]any{"email": "dillon@example.com"}).Serialize()
	if err != nil {
		t.Fatal(err)
	}
	if w := f.do(raw); w.Code != http.StatusUnauthorized {
		t.Errorf("HS256 token: status = %d, want 401", w.Code)
	}
}

func TestAccessAuthFailsClosed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	run := func(cfg AccessConfig) (*httptest.ResponseRecorder, error) {
		mw, err := NewAccessAuth(context.Background(), cfg)
		if err != nil {
			return nil, err
		}
		r := gin.New()
		r.GET("/admin", mw, func(c *gin.Context) { c.String(http.StatusOK, AdminEmail(c)) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		return w, nil
	}

	t.Run("unconfigured: admin disabled", func(t *testing.T) {
		w, err := run(AccessConfig{})
		if err != nil || w.Code != http.StatusServiceUnavailable {
			t.Fatalf("got status %v err %v, want 503", w, err)
		}
	})
	t.Run("dev bypass without Access config", func(t *testing.T) {
		w, err := run(AccessConfig{DevBypass: true})
		if err != nil || w.Code != http.StatusOK || w.Body.String() != devAdminEmail {
			t.Fatalf("got %d %q err %v", w.Code, w.Body.String(), err)
		}
	})
	t.Run("dev bypass is ignored when Access is configured", func(t *testing.T) {
		f := newAccessFixture(t, "dillon@example.com")
		_ = f // fixture constructs the verifier path; now assert bypass flag can't open it
		mw, err := NewAccessAuth(context.Background(), AccessConfig{
			TeamDomain: f.srv.URL, Audience: testAUD, AllowedEmails: []string{"dillon@example.com"}, DevBypass: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		r := gin.New()
		r.GET("/admin", mw, func(c *gin.Context) { c.String(http.StatusOK, "reached") })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/admin", nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 (bypass must not apply when Access is configured)", w.Code)
		}
	})
	t.Run("partial config is a startup error", func(t *testing.T) {
		if _, err := run(AccessConfig{TeamDomain: "https://x.cloudflareaccess.com"}); err == nil {
			t.Fatal("expected error for incomplete Access config")
		}
		if _, err := run(AccessConfig{Audience: "aud", AllowedEmails: []string{"a@b.c"}}); err == nil {
			t.Fatal("expected error when team domain is missing")
		}
	})
}

func TestRequireSameOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	build := func(loopback bool) *gin.Engine {
		r := gin.New()
		r.Use(RequireSameOrigin([]string{"https://dillonthompson.com"}, loopback))
		r.Any("/x", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		return r
	}
	do := func(r *gin.Engine, method string, hdr map[string]string, body string) int {
		req := httptest.NewRequest(method, "/x", nil)
		if body != "" {
			req = httptest.NewRequest(method, "/x", stringsReader(body))
		}
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}
	json := map[string]string{"Content-Type": "application/json"}

	r := build(false)
	if got := do(r, http.MethodGet, map[string]string{"Sec-Fetch-Site": "cross-site"}, ""); got != http.StatusNoContent {
		t.Errorf("GET should pass through, got %d", got)
	}
	if got := do(r, http.MethodPost, merge(json, map[string]string{"Origin": "https://dillonthompson.com", "Sec-Fetch-Site": "same-origin"}), `{}`); got != http.StatusNoContent {
		t.Errorf("same-origin POST: got %d", got)
	}
	if got := do(r, http.MethodPost, merge(json, map[string]string{"Origin": "https://evil.example"}), `{}`); got != http.StatusForbidden {
		t.Errorf("cross-origin POST: got %d, want 403", got)
	}
	if got := do(r, http.MethodPost, merge(json, map[string]string{"Sec-Fetch-Site": "cross-site"}), `{}`); got != http.StatusForbidden {
		t.Errorf("cross-site POST: got %d, want 403", got)
	}
	if got := do(r, http.MethodPost, map[string]string{"Content-Type": "application/x-www-form-urlencoded"}, `a=b`); got != http.StatusUnsupportedMediaType {
		t.Errorf("form POST: got %d, want 415", got)
	}
	if got := do(r, http.MethodDelete, map[string]string{"Origin": "https://evil.example"}, ""); got != http.StatusForbidden {
		t.Errorf("cross-origin DELETE: got %d, want 403", got)
	}
	if got := do(r, http.MethodPost, merge(json, map[string]string{"Origin": "http://localhost:5173"}), `{}`); got != http.StatusForbidden {
		t.Errorf("localhost origin must be rejected outside dev: got %d", got)
	}
	if got := do(build(true), http.MethodPost, merge(json, map[string]string{"Origin": "http://localhost:5173"}), `{}`); got != http.StatusNoContent {
		t.Errorf("localhost origin should pass in dev: got %d", got)
	}
}

func stringsReader(s string) *strings.Reader { return strings.NewReader(s) }

func merge(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}
