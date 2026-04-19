terraform {
  required_version = ">= 1.6"

  required_providers {
    hcloud = {
      source  = "hetznercloud/hcloud"
      version = "~> 1.50"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      # Provider v5 rewrote resources to match the Cloudflare API v4 shape.
      # Key differences vs v4: cloudflare_record → cloudflare_dns_record,
      # rulesets use list-of-objects syntax, and "value" → "content".
      version = "~> 5.0"
    }
  }

  # Local state for a personal project. If this ever goes multi-maintainer,
  # swap to a remote backend (S3 + encryption, Terraform Cloud, etc.).
  # The state file will contain secrets — keep it out of git (.gitignore).
  backend "local" {
    path = "terraform.tfstate"
  }
}
