# /eigen-deploy - Deploy Web App to Eigen Server

Deploy a Vite web application to the Eigen server with nginx and SSL.

## Usage

```
/eigen-deploy <domain>
```

**Example:** `/eigen-deploy silicon.livingbinaries.com`

## Arguments

- `$ARGUMENTS` — The domain name to deploy to (required)

## Server

- **Host:** `josh@84.32.220.106`
- **Sites root:** `~/sites/`
- **Web server:** nginx
- **SSL:** Let's Encrypt (certbot)

---

## PROTOCOL

You are deploying the current project's web build to the Eigen server. Execute each phase sequentially, displaying clear status for every step. Stop and report if any phase fails.

### Phase 0: Validate Inputs

Extract the domain from `$ARGUMENTS`. If no domain is provided, stop and ask for one.

Set these variables for the rest of the protocol:

```
DOMAIN = first argument from $ARGUMENTS
REMOTE = josh@84.32.220.106
SITE_NAME = domain with dots replaced by hyphens (e.g., silicon-livingbinaries-com)
REMOTE_PATH = ~/sites/${SITE_NAME}
```

Display:
```
EIGEN DEPLOY
Domain:      ${DOMAIN}
Site name:   ${SITE_NAME}
Remote path: ${REMOTE_PATH}
```

### Phase 1: Build for Production

Run the project's Vite build:

```bash
npm run build
```

Verify `dist/` exists and contains `index.html`. Report the file count and total size.

Display:
```
[1/6] BUILD ........ OK (N files, N KB)
```

### Phase 2: Copy to Remote

Use `rsync` to deploy the built files. This avoids re-uploading unchanged assets:

```bash
rsync -avz --delete dist/ josh@84.32.220.106:~/sites/${SITE_NAME}/
```

If rsync fails due to SSH authentication, try `scp -r` as fallback. If both fail, inform the user about the SSH authentication issue and suggest solutions:
- SSH key setup: `ssh-copy-id josh@84.32.220.106`
- Or use `sshpass` if password-based auth is needed

Display:
```
[2/6] UPLOAD ....... OK (synced to ~/sites/${SITE_NAME}/)
```

### Phase 3: Verify Remote Files

Confirm the files landed correctly:

```bash
ssh josh@84.32.220.106 "ls -la ~/sites/${SITE_NAME}/ && cat ~/sites/${SITE_NAME}/index.html | head -5"
```

Display:
```
[3/6] VERIFY ....... OK (index.html present)
```

### Phase 4: Configure nginx

Generate and install an nginx site configuration. First check if a config already exists:

```bash
ssh josh@84.32.220.106 "cat /etc/nginx/sites-available/${SITE_NAME} 2>/dev/null || echo 'NO_EXISTING_CONFIG'"
```

If no config exists (or needs updating), create it:

```bash
ssh josh@84.32.220.106 "sudo tee /etc/nginx/sites-available/${SITE_NAME}" << 'NGINX_EOF'
server {
    listen 80;
    listen [::]:80;
    server_name ${DOMAIN};

    root /home/josh/sites/${SITE_NAME};
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
}
NGINX_EOF
```

Enable the site and test nginx:

```bash
ssh josh@84.32.220.106 "sudo ln -sf /etc/nginx/sites-available/${SITE_NAME} /etc/nginx/sites-enabled/${SITE_NAME} && sudo nginx -t && sudo systemctl reload nginx"
```

If `sudo` requires a password, inform the user. Possible solutions:
- The user can add password-less sudo for nginx commands
- Or provide the password interactively via `ssh -t` for a TTY

Display:
```
[4/6] NGINX ........ OK (site enabled, config valid)
```

### Phase 5: SSL Certificate

Request a Let's Encrypt certificate via certbot:

```bash
ssh josh@84.32.220.106 "sudo certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos --email josh@livingbinaries.com --redirect"
```

The `--redirect` flag automatically updates the nginx config to redirect HTTP to HTTPS.

If certbot is not installed:
```bash
ssh josh@84.32.220.106 "sudo apt-get update && sudo apt-get install -y certbot python3-certbot-nginx && sudo certbot --nginx -d ${DOMAIN} --non-interactive --agree-tos --email josh@livingbinaries.com --redirect"
```

If this fails due to DNS not pointing to the server yet, warn the user:
```
WARNING: SSL failed. Make sure DNS for ${DOMAIN} points to 84.32.220.106
You can add an A record and re-run this phase, or run manually:
  ssh josh@84.32.220.106 "sudo certbot --nginx -d ${DOMAIN}"
```

Display:
```
[5/6] SSL .......... OK (Let's Encrypt certificate issued)
```

### Phase 6: Final Verification

Test that the site is reachable:

```bash
curl -sI "https://${DOMAIN}" | head -10
```

Display the final summary:

```
[6/6] LIVE ......... OK

=========================================
  DEPLOYED: https://${DOMAIN}
=========================================

  Server:   84.32.220.106
  Path:     ~/sites/${SITE_NAME}/
  nginx:    /etc/nginx/sites-available/${SITE_NAME}
  SSL:      Let's Encrypt (auto-renew)

  To redeploy:  /eigen-deploy ${DOMAIN}
  To remove:    ssh josh@84.32.220.106 "sudo rm /etc/nginx/sites-enabled/${SITE_NAME} && sudo nginx -t && sudo systemctl reload nginx"
```

---

## SSH Authentication Notes

This command requires SSH access to `josh@84.32.220.106`. If password authentication is needed:

1. **Best option:** Set up SSH key auth once:
   ```bash
   ssh-copy-id josh@84.32.220.106
   ```

2. **For sudo commands:** The nginx and certbot steps require `sudo`. If the remote user needs a password for sudo, those commands will need `-t` flag for TTY allocation, or the user can configure passwordless sudo for specific commands.

3. **Fallback:** If SSH keys aren't set up, each SSH/rsync command will prompt for a password. This works but is tedious for 6+ commands.

---

## Error Recovery

- **Build fails:** Fix the build errors locally first, then re-run
- **Upload fails:** Check SSH connectivity: `ssh josh@84.32.220.106 whoami`
- **nginx fails:** Check config syntax: `ssh josh@84.32.220.106 "sudo nginx -t"`
- **SSL fails:** Verify DNS A record points to 84.32.220.106, then retry certbot manually
- **Site not loading:** Check nginx error log: `ssh josh@84.32.220.106 "sudo tail -20 /var/log/nginx/error.log"`
