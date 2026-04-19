variable "hcloud_token" {
  description = "Hetzner Cloud API token with read/write on Project → API Tokens"
  type        = string
  sensitive   = true
}

variable "cloudflare_api_token" {
  description = "Cloudflare API token scoped to Zone.DNS (Edit) for dillonthompson.com"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Cloudflare zone ID for dillonthompson.com"
  type        = string
}

variable "server_type" {
  description = "Hetzner instance type. cx22 = 2 vCPU / 4 GB / 40 GB (~$4.50/mo)"
  type        = string
  default     = "cx22"
}

variable "location" {
  description = "Hetzner datacenter (ash = Ashburn VA, fsn1 = Falkenstein DE, hel1 = Helsinki FI, hil = Hillsboro OR, nbg1 = Nuremberg DE)"
  type        = string
  default     = "ash"
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
