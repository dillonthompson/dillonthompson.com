output "server_ipv4" {
  description = "Public IPv4 address of the server"
  value       = hcloud_server.web.ipv4_address
}

output "server_ipv6" {
  description = "Public IPv6 address of the server"
  value       = hcloud_server.web.ipv6_address
}

output "server_name" {
  description = "Hetzner-reported server name (shows up in Hetzner console)"
  value       = hcloud_server.web.name
}

output "ssh_cmd" {
  description = "One-liner to SSH in as the personal user"
  value       = "ssh dillon@${hcloud_server.web.ipv4_address}"
}

output "ssh_keyscan_cmd" {
  description = "Run this after first boot to capture the SSH host key for GitHub Actions secret SSH_KNOWN_HOSTS"
  value       = "ssh-keyscan -t ed25519 ${hcloud_server.web.ipv4_address}"
}
