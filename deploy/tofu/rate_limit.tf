# Cloudflare rate limiting for /api/* endpoints.
#
# Free plan allowance: ONE rate-limit rule per zone. Pro = 5, Business = 10.
# If you upgrade to Pro, split this into a broader /api/* rule and a stricter
# rule for the DB-heavy endpoints (profile, experience).
#
# Rules fire per-client-IP, per-colo. `block` action returns 429 at the edge
# without ever reaching the origin, so even a flood is cheap for us.

resource "cloudflare_ruleset" "rate_limit" {
  zone_id     = var.cloudflare_zone_id
  name        = "rate-limit-api"
  description = "Protect /api/* from traffic spikes"
  kind        = "zone"
  phase       = "http_ratelimit"

  rules = [
    {
      action      = "block"
      description = "60 req/min per IP to /api/*"
      enabled     = true
      expression  = "(starts_with(http.request.uri.path, \"/api/\"))"

      ratelimit = {
        # cf.colo.id + ip.src means the 60/min budget is counted per Cloudflare
        # data center. Globally a single IP could technically exceed 60/min if
        # they bounce between colos, but in practice this covers realistic abuse
        # cases while keeping the rule simple.
        characteristics = ["cf.colo.id", "ip.src"]

        # A regular visitor loading the homepage fires ~2-3 /api calls (profile +
        # experience). 60/min leaves a big margin for refresh-happy users while
        # still cutting off crawlers and scripted abuse.
        period              = 60
        requests_per_period = 60

        # After tripping the limit, block further requests for 60s. Too short and
        # they just retry; too long and legitimate users get stuck behind an IP
        # that ran a bad script 10 minutes ago.
        mitigation_timeout = 60
        requests_to_origin = false
      }
    }
  ]
}
