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
        # cf.colo.id + ip.src means the budget is counted per Cloudflare data
        # center per source IP. Globally a single IP could technically double
        # this by bouncing between colos, but in practice this covers realistic
        # abuse cases while keeping the rule simple.
        characteristics = ["cf.colo.id", "ip.src"]

        # Free Cloudflare plan restricts `period` to 10 seconds. Pro+ unlocks
        # longer windows. 10 req / 10s ≈ 60 req/min for steady traffic, and a
        # visitor loading the homepage only fires 2-3 /api calls, so bursts
        # from real users stay well under the cap.
        period              = 10
        requests_per_period = 10

        # mitigation_timeout: how long a tripped IP stays blocked. On free plan
        # this is also capped, so match `period` to stay safely within limits.
        mitigation_timeout = 10
        requests_to_origin = false
      }
    }
  ]
}
