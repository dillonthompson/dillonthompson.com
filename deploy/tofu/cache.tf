# Edge caching for the server-rendered blog.
#
# Cloudflare does not cache HTML by default, so without this rule the
# `Cache-Control: s-maxage=60` the Go API sends is ignored and every blog
# request reaches the origin (and Neon). "Respect origin" lets the API stay in
# charge of TTLs: a newly published (or unpublished) post is reflected within
# ~2 minutes, with no purge step.
#
# Scope is deliberately narrow — only public blog URLs. /admin and
# /api/v1/admin are never matched, and the admin API sends `no-store` anyway.
#
# Free plan: Cache Rules are included (10 rules per zone).

resource "cloudflare_ruleset" "blog_cache" {
  zone_id     = var.cloudflare_zone_id
  name        = "cache-blog"
  description = "Cache public blog pages, feed and sitemap per origin Cache-Control"
  kind        = "zone"
  phase       = "http_request_cache_settings"

  rules = [
    {
      action      = "set_cache_settings"
      description = "Cache /blog*, /rss.xml, /sitemap.xml (respect origin TTLs)"
      enabled     = true
      expression  = "(http.request.uri.path eq \"/blog\") or starts_with(http.request.uri.path, \"/blog/\") or (http.request.uri.path eq \"/rss.xml\") or (http.request.uri.path eq \"/sitemap.xml\")"

      action_parameters = {
        cache = true
        edge_ttl = {
          mode = "respect_origin"
        }
        browser_ttl = {
          mode = "respect_origin"
        }
      }
    }
  ]
}
