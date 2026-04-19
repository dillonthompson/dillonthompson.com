locals {
  # Render cloud-init with the SSH keys and systemd unit inlined. Using templatefile
  # instead of cloud-init's built-in fetching means tofu owns the whole config —
  # state drift between git and the running server shows up as a diff on `plan`.
  cloud_init = templatefile("${path.module}/../cloud-init.yml", {
    ssh_public_key        = var.ssh_public_key
    deploy_ssh_public_key = var.deploy_ssh_public_key
    # Indent by 6 spaces so the unit file sits cleanly inside the cloud-init YAML's
    # write_files content block.
    systemd_unit = replace(
      file("${path.module}/../dillonthompson.service"),
      "\n",
      "\n      "
    )
  })
}

# --- Hetzner firewall --------------------------------------------------------
# Cloud-init also enables ufw, so this is defense-in-depth. Hetzner's firewall
# runs at the hypervisor layer so traffic is blocked before it reaches the VM.

resource "hcloud_firewall" "web" {
  name = "dillonthompson-web"

  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "22"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "SSH"
  }

  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "80"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "HTTP (for Let's Encrypt challenges and HTTP->HTTPS redirect)"
  }

  rule {
    direction = "in"
    protocol  = "tcp"
    port      = "443"
    source_ips = ["0.0.0.0/0", "::/0"]
    description = "HTTPS"
  }

  rule {
    direction   = "in"
    protocol    = "udp"
    port        = "443"
    source_ips  = ["0.0.0.0/0", "::/0"]
    description = "HTTP/3 (QUIC)"
  }
}

# --- SSH key registered with Hetzner ----------------------------------------
# Hetzner injects this into /root/.ssh/authorized_keys at boot. Cloud-init then
# adds the same key to the `dillon` user below.

resource "hcloud_ssh_key" "personal" {
  name       = "dillon-personal"
  public_key = var.ssh_public_key
}

# --- The server -------------------------------------------------------------

resource "hcloud_server" "web" {
  name         = "dillonthompson"
  server_type  = var.server_type
  image        = var.image
  location     = var.location
  ssh_keys     = [hcloud_ssh_key.personal.id]
  firewall_ids = [hcloud_firewall.web.id]
  user_data    = local.cloud_init

  public_net {
    ipv4_enabled = true
    ipv6_enabled = true
  }

  # If you rotate cloud-init or the systemd unit, a replace will recreate the
  # server — destructive for Caddy's ACME cert cache. For changes post-launch,
  # update /etc/systemd/system/dillonthompson.service over SSH manually and
  # reload daemon, instead of letting tofu replace the whole box.
  lifecycle {
    ignore_changes = [user_data]
  }
}

# --- Cloudflare DNS ---------------------------------------------------------
# A record for the apex, CNAME for www. Proxy (orange cloud) is off by default
# so Let's Encrypt HTTP-01 challenges work. Flip `cloudflare_proxy = true` once
# HTTPS is verified; Caddy's cert will survive behind Cloudflare's proxy.

resource "cloudflare_dns_record" "apex" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain
  type    = "A"
  content = hcloud_server.web.ipv4_address
  proxied = var.cloudflare_proxy
  # ttl: 1 means "auto" (only valid when proxied through Cloudflare).
  ttl = var.cloudflare_proxy ? 1 : 300
}

resource "cloudflare_dns_record" "www" {
  zone_id = var.cloudflare_zone_id
  name    = "www.${var.domain}"
  type    = "CNAME"
  content = var.domain
  proxied = var.cloudflare_proxy
  ttl     = var.cloudflare_proxy ? 1 : 300
}

resource "cloudflare_dns_record" "apex_aaaa" {
  zone_id = var.cloudflare_zone_id
  name    = var.domain
  type    = "AAAA"
  content = hcloud_server.web.ipv6_address
  proxied = var.cloudflare_proxy
  ttl     = var.cloudflare_proxy ? 1 : 300
}
