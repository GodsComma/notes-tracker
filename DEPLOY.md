# Deployment Guide

> Self-hosted production deployment with HTTPS (self-signed certificate).  
> Designed for a Linux ARM64 host on a home network or private VPN.

---

## Prerequisites

**On your build machine:**
- Go 1.21+ (`go version`); the `go 1.25` toolchain declared in `go.mod` is fetched automatically
- A C compiler (macOS: `xcode-select --install`) — required by the SQLite driver
- Node.js 18+ and npm (`node --version`)
- `rsync` and `ssh`

**On the deploy host:**
- Linux (Debian/Ubuntu tested; any systemd distro works)
- `openssl` (`openssl version`)
- `sqlite3` (optional, for manual DB inspection)
- SSH access with `sudo` privileges

---

## Overview

There are two scripts:

| Script | Runs on | When |
|---|---|---|
| `script/provision-host` | **Deploy host** | Once, before the first deploy |
| `script/deploy` | **Build machine** | Every deploy (initial and updates) |

The sequence for a first deployment is:

```
1. SSH into the host → run provision-host
2. On your build machine → run script/deploy
3. On each client device → trust the certificate
```

Updates after that are just `script/deploy`.

---

## Step 1 — Host Setup (run once on the server)

SSH into your deploy host, then copy and run the setup script.

**Transfer the script to the host:**

```bash
scp script/provision-host user@192.168.1.X:~/provision-host
ssh -t user@192.168.1.X 'chmod +x ~/provision-host && sudo ~/provision-host'
```

The script will prompt you for:
- **Hostname** — the domain or hostname browsers will use (e.g. `notes.internal`)
- **Server IP** — the host's LAN IP address (e.g. `192.168.1.50`)
- **Deploy user** — your SSH login user (so they can write to `/opt/notes/app`)

It will then:
1. Create a `notes` system user (no shell, no home directory)
2. Create the directory structure under `/opt/notes/`
3. Generate a self-signed TLS certificate valid for 10 years
4. Generate a `JWT_SECRET` and write the environment file
5. Install and enable the `notes.service` systemd unit
6. Print the certificate file path you'll need for client trust

> The service won't start yet — there's no binary to run until Step 2.

---

## Step 2 — Build and Deploy (run from your build machine)

**Configure the target** — set these in `.env` at the repo root (`script/bootstrap` creates it from `.env.example`; it is gitignored):

```bash
DEPLOY_USER=pi                 # Your SSH login on the host
DEPLOY_HOST=192.168.1.X        # Host IP or hostname
DEPLOY_APP_DIR=/opt/notes/app  # Default; must match APP_DIR in provision-host
```

**Run:**

```bash
script/deploy
```

This will:
1. Install/verify dependencies (`script/bootstrap`)
2. Build the Vue frontend (`npm run build`)
3. Cross-compile the Go binary for Linux ARM64
4. Bundle everything into `dist/`
5. `rsync` the bundle to the host
6. Restart the `notes` service via SSH
7. Print the service status

The app is now running at `https://<your-hostname>`.

> **First deploy only:** the browser will show a certificate error until you complete Step 3.

---

## Step 3 — Trust the Certificate (once per device)

The TLS certificate is self-signed, so each device that accesses the server needs to trust it.

First, copy the certificate from the host to your local machine:

```bash
scp user@192.168.1.X:/opt/notes/certs/cert.pem ~/notes-cert.pem
```

Then install it on each device:

### macOS
```bash
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain ~/notes-cert.pem
```
Then quit and reopen any browsers.

### iOS / iPadOS
1. AirDrop or email `cert.pem` to the device
2. Tap the file → **Settings → General → VPN & Device Management** → tap the profile → Install
3. **Settings → General → About → Certificate Trust Settings** → enable full trust for the certificate

### Linux (Debian / Ubuntu)
```bash
sudo cp ~/notes-cert.pem /usr/local/share/ca-certificates/notes.crt
sudo update-ca-certificates
```

### Windows
1. Double-click `cert.pem`
2. Click **Install Certificate**
3. Choose **Local Machine** → **Place all certificates in the following store**
4. Browse → **Trusted Root Certification Authorities** → OK → Finish

---

## Updating the App

After any code change, just run the deploy script again from your build machine:

```bash
script/deploy
```

Database migrations run automatically on startup — new migration files are applied and already-applied ones are skipped. No manual DB work needed.

---

## Database Console

```bash
script/console production          # read-only sqlite3 on the host's DB
script/console production --write  # writable (asks for confirmation)
```

---

## Checking Logs

```bash
ssh user@192.168.1.X 'journalctl -u notes -f'
```

To see the last 100 lines:

```bash
ssh user@192.168.1.X 'journalctl -u notes -n 100 --no-pager'
```

---

## Backup

Persistent state lives entirely in `/opt/notes/app/data/`. A daily cron job is set up by `provision-host` that tarballs this directory to `/opt/notes/backups/` and prunes backups older than 30 days.

To restore from a backup:

```bash
sudo systemctl stop notes
sudo tar -xzf /opt/notes/backups/notes-YYYYMMDD.tar.gz -C /opt/notes/app
sudo systemctl start notes
```

---

## Certificate Renewal

Self-signed certificates don't auto-renew. The setup script generates a 10-year cert, so you won't need to do this often. When the time comes:

```bash
ssh user@192.168.1.X 'sudo /opt/notes/renew-cert.sh'
```

`provision-host` installs this renewal script automatically. After renewal you'll need to re-trust the new certificate on each device (Step 3 again).

---

## Environment Variables Reference

All variables live in `/opt/notes/env` on the host. Edit with:

```bash
ssh user@192.168.1.X 'sudo nano /opt/notes/env && sudo systemctl restart notes'
```

| Variable | Required | Default | Description |
|---|---|---|---|
| `JWT_SECRET` | Yes | — | 32-byte hex secret for signing auth tokens. Change this and all users are logged out. |
| `TLS_CERT` | Yes | — | Absolute path to PEM certificate |
| `TLS_KEY` | Yes | — | Absolute path to PEM private key |
| `TLS_PORT` | No | `443` | HTTPS listen port |
| `HTTP_PORT` | No | `80` | HTTP→HTTPS redirect port. Leave empty to disable. |
| `CORS_ORIGINS` | No | `http://localhost:5173` | Comma-separated list of allowed origins |
| `APP_ENV` | No | — | Set to `production` to enable Secure flag on auth cookies |
| `DATA_DIR` | No | `data` | Path to the runtime data directory |
| `PORT` | No | `3001` | Plaintext HTTP port (only used when TLS is not configured) |

---

## Troubleshooting

**Service won't start**
```bash
ssh user@192.168.1.X 'journalctl -u notes -n 50 --no-pager'
```
Common causes: missing env file, wrong cert path, port already in use.

**Browser shows certificate error after trusting**  
Quit the browser fully and reopen. On macOS, check Keychain Access to confirm the cert is listed under "System" with "Always Trust".

**`rsync` permission denied**  
Ensure your SSH user is in the same group as the `notes` app directory:
```bash
ssh user@192.168.1.X 'sudo chown -R $USER:$USER /opt/notes/app'
```

**`JWT_SECRET` environment variable is required**  
The env file is missing or the path in the systemd unit is wrong. Check:
```bash
ssh user@192.168.1.X 'sudo cat /opt/notes/env && sudo systemctl cat notes'
```

**Port 443 / 80 already in use**  
Another process (nginx, apache) is listening. Either stop it or change `TLS_PORT` / `HTTP_PORT` in the env file to unused ports and use nftables/iptables to redirect 443→new port.
