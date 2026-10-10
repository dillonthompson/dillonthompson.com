# Cloudflare Access in front of the blog admin.
#
# Access gates /admin* (the editor SPA route) and /api/v1/admin* (the API) at
# Cloudflare's edge: unauthenticated visitors get a login page and never reach
# the origin. Login is a one-time code emailed to var.admin_email.
#
# The Go API does NOT rely on this alone. It verifies the Access JWT on every
# admin request (signature, issuer, audience, expiry, email allow-list), so a
# request that reaches the origin directly is still rejected. That needs the
# values from this module's outputs passed to the container:
#   CF_ACCESS_TEAM_DOMAIN, CF_ACCESS_AUD, ADMIN_EMAILS  (see deploy/README.md)
#
# Prerequisite (one-time, dashboard): create your Zero Trust organization
# (Zero Trust → Settings → Custom pages / team name) so var.access_team_name
# exists. The Free plan covers up to 50 users; the one-time-PIN login method is
# enabled by default.

resource "cloudflare_zero_trust_access_policy" "blog_admin" {
  account_id       = var.cloudflare_account_id
  name             = "blog-admin (${var.domain})"
  decision         = "allow"
  session_duration = "24h"

  include = [
    {
      email = {
        email = var.admin_email
      }
    }
  ]
}

resource "cloudflare_zero_trust_access_application" "blog_admin" {
  account_id       = var.cloudflare_account_id
  name             = "${var.domain} blog admin"
  type             = "self_hosted"
  session_duration = "24h"

  # Don't list this in the Access App Launcher; it's reached by URL.
  app_launcher_visible = false

  destinations = [
    { type = "public", uri = "${var.domain}/admin*" },
    { type = "public", uri = "${var.domain}/api/v1/admin*" },
  ]

  policies = [
    {
      id         = cloudflare_zero_trust_access_policy.blog_admin.id
      precedence = 1
    }
  ]
}
