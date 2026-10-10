variable "hcloud_token" {
  description = "Hetzner Cloud API token with read/write on Project → API Tokens"
  type        = string
  sensitive   = true
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token. Needs Zone: DNS Edit, Zone WAF Edit, Zone Read, Cache Rules Edit (zone) and Account: Access: Apps and Policies Edit — see terraform.tfvars.example"
  type        = string
  sensitive   = true
}

variable "cloudflare_account_id" {
  description = "Cloudflare account ID (needed for Zero Trust / Access). Dashboard → account home → ⋯ menu → Copy account ID."
  type        = string
}

variable "access_team_name" {
  description = "Your Zero Trust team name — the <name> in <name>.cloudflareaccess.com. Create the organization in the dashboard first."
  type        = string
}

variable "admin_email" {
  description = "The one email address allowed into the blog admin (Cloudflare Access policy and the API's allow-list)"
  type        = string
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for dillonthompson.com"
  type        = string
}

variable "server_type" {
  description = <<-EOT
    Hetzner instance type. Server families are location-bound — check the Hetzner
    console for what's available where you're deploying.
      - US (ash, hil): cpx11 (2 vCPU / 2 GB / 40 GB, ~$4.79/mo) — this default
      - EU (fsn1, nbg1, hel1): cx23 (2 vCPU / 4 GB / 40 GB, ~€3.79/mo)
    ARM (cax*) is also cheap but our image is linux/amd64 only.
  EOT
  type    = string
  default = "cpx11"
}

variable "location" {
  description = "Hetzner datacenter. ash = Ashburn VA, hil = Hillsboro OR, fsn1 = Falkenstein DE, nbg1 = Nuremberg DE, hel1 = Helsinki FI. Must match server_type availability."
  type        = string
  default     = "hil"
}

variable "image" {
  description = "Hetzner OS image. Ubuntu 24.04 LTS is current LTS at time of writing."
  type        = string
  default     = "ubuntu-24.04"
}

variable "domain" {
  description = "Root domain this server serves"
  type        = string
  default     = "dillonthompson.com"
}

variable "ssh_public_key" {
  description = "Your personal SSH public key for manual access (contents, not path)"
  type        = string
}

variable "deploy_ssh_public_key" {
  description = "Deploy SSH public key — corresponds to SSH_PRIVATE_KEY in GitHub Secrets"
  type        = string
}

variable "cloudflare_proxy" {
  description = "Proxy traffic through Cloudflare (orange cloud). Set false for the initial Let's Encrypt cert issuance, flip to true once HTTPS is live."
  type        = bool
  default     = false
}
