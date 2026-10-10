package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

// AdminEmailKey is the gin context key holding the authenticated admin's email.
const AdminEmailKey = "admin_email"

// accessJWTHeader is set by Cloudflare Access on every request it lets through.
const accessJWTHeader = "Cf-Access-Jwt-Assertion"

// devAdminEmail is the identity used when the local-dev bypass is active.
const devAdminEmail = "dev@localhost"

// AccessConfig configures admin authentication.
//
// Production relies on Cloudflare Access: it gates /admin* and /api/v1/admin/*
// at the edge, and this middleware independently verifies the JWT it attaches,
// so a request that reaches the origin directly (leaked origin IP, a
// misconfigured Access app) is still rejected.
type AccessConfig struct {
	// TeamDomain is the Zero Trust team URL, e.g. https://you.cloudflareaccess.com.
	TeamDomain string
	// Audience is the Access application's AUD tag.
	Audience string
	// AllowedEmails lists who may use the admin, matched case-insensitively.
	// Enforced here in addition to the Access policy.
	AllowedEmails []string
	// DevBypass skips verification and authenticates as devAdminEmail. It is
	// honoured ONLY when no Access settings are configured at all, so a
	// production deploy (which must configure Access for the admin to work)
	// can never run with the bypass active.
	DevBypass bool
}

func (c AccessConfig) accessConfigured() bool {
	return c.TeamDomain != "" || c.Audience != "" || len(c.AllowedEmails) > 0
}

// BypassActive reports whether the dev bypass will actually be in effect.
func (c AccessConfig) BypassActive() bool {
	return c.DevBypass && !c.accessConfigured()
}

// NewAccessAuth builds the admin authentication middleware. It fails closed:
// with no configuration the admin is disabled, and with partial configuration
// construction returns an error.
func NewAccessAuth(ctx context.Context, cfg AccessConfig) (gin.HandlerFunc, error) {
	if !cfg.accessConfigured() {
		if cfg.DevBypass {
			slog.Warn("ADMIN_DEV_BYPASS is on: admin endpoints are UNAUTHENTICATED. Local development only.")
			return func(c *gin.Context) {
				c.Set(AdminEmailKey, devAdminEmail)
				c.Next()
			}, nil
		}
		slog.Warn("admin disabled: set CF_ACCESS_TEAM_DOMAIN, CF_ACCESS_AUD and ADMIN_EMAILS to enable it")
		return func(c *gin.Context) {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "admin is not configured"})
		}, nil
	}

	if cfg.DevBypass {
		slog.Warn("ADMIN_DEV_BYPASS ignored: Cloudflare Access is configured")
	}
	if cfg.TeamDomain == "" || cfg.Audience == "" || len(cfg.AllowedEmails) == 0 {
		return nil, fmt.Errorf("incomplete Cloudflare Access config: CF_ACCESS_TEAM_DOMAIN, CF_ACCESS_AUD and ADMIN_EMAILS must all be set")
	}

	teamDomain, err := normalizeTeamDomain(cfg.TeamDomain)
	if err != nil {
		return nil, err
	}

	// The remote key set fetches and caches Cloudflare's signing keys, and
	// refreshes them when it sees an unknown key ID (rotation).
	keySet := oidc.NewRemoteKeySet(ctx, teamDomain+"/cdn-cgi/access/certs")
	verifier := oidc.NewVerifier(teamDomain, keySet, &oidc.Config{
		// Checks signature, expiry, issuer and that Audience is in `aud`.
		ClientID: cfg.Audience,
		// Pin the algorithm: Access signs with RS256.
		SupportedSigningAlgs: []string{oidc.RS256},
	})

	allowed := make(map[string]struct{}, len(cfg.AllowedEmails))
	for _, e := range cfg.AllowedEmails {
		allowed[strings.ToLower(strings.TrimSpace(e))] = struct{}{}
	}

	return func(c *gin.Context) {
		raw := c.GetHeader(accessJWTHeader)
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		token, err := verifier.Verify(c.Request.Context(), raw)
		if err != nil {
			slog.Warn("admin auth: token rejected", "error", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		var claims struct {
			Email string `json:"email"`
		}
		if err := token.Claims(&claims); err != nil || claims.Email == "" {
			slog.Warn("admin auth: token has no email claim")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		email := strings.ToLower(claims.Email)
		if _, ok := allowed[email]; !ok {
			slog.Warn("admin auth: email not allowed", "email", email)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Set(AdminEmailKey, email)
		c.Next()
	}, nil
}

func normalizeTeamDomain(raw string) (string, error) {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("invalid CF_ACCESS_TEAM_DOMAIN %q", raw)
	}
	return u.Scheme + "://" + u.Host, nil
}

// AdminEmail returns the authenticated admin's email from the gin context.
func AdminEmail(c *gin.Context) string {
	return c.GetString(AdminEmailKey)
}
