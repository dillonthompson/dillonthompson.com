# Deploying dillonthompson.com

A single-container deploy of the full site (Caddy + Go API + React bundle) onto a Hetzner VPS.
Managed declaratively with OpenTofu, provisioned via cloud-init, and deployed via GitHub Actions.

```
GitHub push to main
  ↓
Actions builds Dockerfile.prod
  ↓
Pushes image to ghcr.io/dillonthompson/dillonthompson.com (by digest)
  ↓
SSHes into Hetzner, writes /opt/dillonthompson/.env, systemctl restart
  ↓
Container pulls new image, starts, serves requests — Caddy retries during restart
```

## One-time setup (before the first deploy)

### 1. Install OpenTofu

```fish
brew install opentofu
```

### 2. Create the Neon database

- Sign up at [neon.tech](https://neon.tech) with your GitHub account
- Create a project called `dillonthompson`
- Grab the connection string from **Dashboard → Connection Details** — save it, you'll paste it into GitHub Secrets in step 5

### 3. Generate a deploy SSH key (separate from your personal one)

```fish
ssh-keygen -t ed25519 -C "dillonthompson-deploy" -f ~/.ssh/dillonthompson-deploy -N ''
```

Keep both files around — public goes into OpenTofu, private goes into a GitHub Secret.

### 4. Fill in OpenTofu variables

```fish
cd deploy/tofu
cp terraform.tfvars.example terraform.tfvars
$EDITOR terraform.tfvars
```

Populate:
- `hcloud_token` — Hetzner Cloud console → Security → API Tokens → Read & Write
- `cloudflare_api_token` — Cloudflare → My Profile → API Tokens → **Create Custom Token** with these permissions, all scoped to the `dillonthompson.com` zone:
  - `Zone → DNS → Edit` (DNS records)
  - `Zone → Zone WAF → Edit` (rate-limit ruleset)
  - `Zone → Zone → Read` (zone lookup)
- `cloudflare_zone_id` — right sidebar of the zone's Cloudflare page
- `ssh_public_key` — your personal `~/.ssh/id_ed25519.pub` contents
- `deploy_ssh_public_key` — contents of `~/.ssh/dillonthompson-deploy.pub`

### 5. Provision the server

```fish
tofu init
tofu apply
```

Grab the outputs — the `server_ipv4` and the `ssh_keyscan_cmd`.

### 6. Capture the SSH host key

You can't know the server's host fingerprint until it exists. Once `tofu apply` finishes, run the keyscan command it printed:

```fish
ssh-keyscan -t ed25519 <server_ipv4>
```

Save that output — it goes into the `SSH_KNOWN_HOSTS` GitHub Secret.

### 7. Populate GitHub Secrets

In your repo → Settings → Secrets and variables → Actions:

| Secret | Value |
|--------|-------|
| `SSH_HOST` | Server IPv4 from `tofu output server_ipv4` |
| `SSH_USER` | `deploy` |
| `SSH_PRIVATE_KEY` | Contents of `~/.ssh/dillonthompson-deploy` (the private key) |
| `SSH_KNOWN_HOSTS` | Output of the `ssh-keyscan` from step 6 |
| `DATABASE_URL` | Neon connection string from step 2 |

Optional repo variables (not secrets, under the same page but on the **Variables** tab):

| Variable | Default if unset |
|----------|------------------|
| `SITE_ADDRESS` | `dillonthompson.com` |
| `LETSENCRYPT_EMAIL` | `dj.thompson715@gmail.com` |

### 8. First deploy

Either push to `main` or trigger the workflow manually from the Actions tab. The workflow builds, pushes to GHCR, SSHes in, and starts the service.

Watch for the health check at the end — first deploy has to resolve Let's Encrypt, which takes ~30 seconds the first time.

### 9. Flip the Cloudflare proxy on (optional)

Once HTTPS is live:

```fish
cd deploy/tofu
tofu apply -var="cloudflare_proxy=true"
```

This puts Cloudflare's orange cloud in front of the origin — DDoS mitigation, analytics, and a CDN for static assets. Caddy's cert still handles origin HTTPS behind it.

## Day-to-day deploys

Push to `main`. That's it.

Roll back by re-running the workflow against an older commit SHA from the Actions tab (workflow_dispatch lets you pick a ref).

## Blog admin (Cloudflare Access)

The blog editor lives at `/admin` and its API at `/api/v1/admin/*`. Two layers
protect it: **Cloudflare Access** gates both paths at the edge (email one-time
code), and the **Go API independently verifies the Access JWT** (signature,
issuer, audience, expiry, and an email allow-list), so a request that reaches the
origin directly is still rejected. With any of the settings missing the admin
stays **disabled** (503) — it never falls back to open.

One-time setup:

1. **Create your Zero Trust organization** in the Cloudflare dashboard (Zero
   Trust → pick a team name). The Free plan covers up to 50 users.
2. **Extend your Cloudflare API token** (Edit token → add):
   `Account → Access: Apps and Policies → Edit` and `Zone → Cache Rules → Edit`.
3. **Add three variables to `terraform.tfvars`** (see `terraform.tfvars.example`):
   `cloudflare_account_id`, `access_team_name`, `admin_email`.
4. **Apply**: `tofu plan`, review, then `tofu apply`. This creates the Access
   application + policy (`access.tf`) and the blog cache rule (`cache.tf`).
5. **Set three GitHub Actions *variables*** (Settings → Secrets and variables →
   Actions → Variables; they aren't secrets) from `tofu output`:
   `CF_ACCESS_TEAM_DOMAIN`, `CF_ACCESS_AUD`, `ADMIN_EMAILS`.
6. **Update the systemd unit on the *existing* server.** Deploys don't rewrite
   it (only cloud-init does, on first boot), so SSH in once:
   ```sh
   scp deploy/dillonthompson.service dillon@<server>:/tmp/
   ssh dillon@<server> 'sudo install -m 644 /tmp/dillonthompson.service /etc/systemd/system/dillonthompson.service && sudo systemctl daemon-reload'
   ```
   The next deploy (or `systemctl restart dillonthompson`) picks up the new `-e` flags.
7. **Deploy**, then visit `https://dillonthompson.com/admin` — you should get
   the Cloudflare login, then the editor.

Notes:
- Order doesn't matter for safety: until steps 5–6 are done the admin just
  returns 503; the public blog works regardless.
- Local dev: set `ADMIN_DEV_BYPASS=true` (see `.env.example`). It is ignored
  whenever the Access settings are present, so it can't be active in prod.
- Public changes show up within ~2 minutes (60s edge cache honoring the API's
  `s-maxage=60`, plus 60s in the browser). That is also how long an unpublished
  or deleted post can linger. There is no stale-while-revalidate window and no
  purge step.
- Rotating the Access application (destroy/recreate) changes `CF_ACCESS_AUD`;
  update the GitHub variable and redeploy.

## File map

| Path | Purpose |
|------|---------|
| `Dockerfile.prod` (repo root) | Multi-stage build: Go binary + React bundle + Caddy runtime |
| `deploy/Caddyfile.prod` | Caddy config baked into the image |
| `deploy/s6-services/` | s6-overlay service definitions for api + caddy |
| `deploy/dillonthompson.service` | systemd unit installed on the server by cloud-init |
| `deploy/cloud-init.yml` | First-boot server bootstrap (installs Docker, creates users, sets up firewall) |
| `deploy/tofu/` | OpenTofu config for Hetzner + Cloudflare |
| `.github/workflows/deploy.yml` | Build → push → deploy pipeline |

## Rotating secrets

- **DB password**: Rotate in Neon, update `DATABASE_URL` GitHub Secret, re-run the workflow
- **Deploy SSH key**: Regenerate, update `deploy_ssh_public_key` in tfvars, `tofu apply`, update `SSH_PRIVATE_KEY` secret

## Debugging on the server

```fish
ssh dillon@<ip>

# Service status
sudo systemctl status dillonthompson

# Live logs
sudo journalctl -u dillonthompson -f

# Inspect running container
sudo docker ps
sudo docker logs dillonthompson

# Manually pull & restart
sudo systemctl restart dillonthompson
```
