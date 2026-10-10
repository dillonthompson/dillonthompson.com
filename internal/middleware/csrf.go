package middleware

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// RequireSameOrigin protects state-changing admin endpoints from cross-site
// requests. The Cloudflare Access session is a cookie, so without this a
// malicious page could make the logged-in admin's browser submit requests.
//
// It rejects unsafe methods when the browser reports a cross-site context
// (Sec-Fetch-Site) or an Origin that isn't ours, and requires a JSON content
// type on bodies (a cross-site form can't send application/json without a CORS
// preflight, which we never answer).
//
// allowedOrigins are full origins, e.g. https://dillonthompson.com. When
// allowLoopback is set (local dev only) http://localhost:* is also accepted,
// since the Vite dev proxy changes the Host header.
func RequireSameOrigin(allowedOrigins []string, allowLoopback bool) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[strings.TrimRight(o, "/")] = struct{}{}
	}

	originOK := func(origin string) bool {
		if _, ok := allowed[origin]; ok {
			return true
		}
		if !allowLoopback {
			return false
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		h := u.Hostname()
		return h == "localhost" || h == "127.0.0.1"
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		if site := c.GetHeader("Sec-Fetch-Site"); site != "" && site != "same-origin" && site != "none" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-site request blocked"})
			return
		}
		if origin := c.GetHeader("Origin"); origin != "" && !originOK(origin) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cross-origin request blocked"})
			return
		}
		if c.Request.ContentLength != 0 && !strings.HasPrefix(c.GetHeader("Content-Type"), "application/json") {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{"error": "content type must be application/json"})
			return
		}
		c.Next()
	}
}
