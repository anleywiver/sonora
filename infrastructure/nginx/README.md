# Nginx + SSL (Sprint 14)

Reverse proxy for the 3 subdomains: `FRONTEND_DOMAIN` (main app),
`ADMIN_DOMAIN`, `API_DOMAIN` (also serves the WebSocket at `/ws`). Same
"sites-available / sites-enabled" idea as classic nginx packaging:
`templates-available/` holds both phases of config, `templates-active/`
(gitignored, one file at a time) is what's actually mounted into the
container and processed by nginx's built-in `envsubst` templating
(official `nginx` image — anything in `/etc/nginx/templates/*.template`
gets `${VAR}`-substituted into `/etc/nginx/conf.d/*.conf` at container
start).

**Butuh input manual dari user**: real domain names pointed at this
VPS's IP (A records for all three subdomains) — nothing here can be
tested end-to-end without them, same class of gap as Google/Bandcamp/
Dropbox credentials elsewhere in this project.

## Bootstrap procedure (first time only)

Certs can't exist before nginx has served the ACME HTTP-01 challenge at
least once, and nginx can't start an HTTPS server block without certs —
so this is a two-phase dance:

1. Fill in `.env`: `FRONTEND_DOMAIN`, `ADMIN_DOMAIN`, `API_DOMAIN` (real
   subdomains, DNS already pointed at this server), `NEXT_PUBLIC_API_URL`
   set to `https://<API_DOMAIN>/api/v1`.
2. Confirm the bootstrap (HTTP-only) template is active — it is by
   default:
   ```
   cp infrastructure/nginx/templates-available/bootstrap.conf.template \
      infrastructure/nginx/templates-active/
   ```
3. Bring everything up: `docker compose up -d`. Nginx now answers plain
   HTTP on all three domains (port 80 only).
4. Request certs (webroot method, shares the `certbot_www` volume nginx
   already serves `/.well-known/acme-challenge/` from):
   ```
   docker compose run --rm certbot certonly --webroot \
     -w /var/www/certbot \
     -d $FRONTEND_DOMAIN -d $ADMIN_DOMAIN -d $API_DOMAIN \
     --email you@example.com --agree-tos --no-eff-email
   ```
5. Switch to the full config and reload:
   ```
   cp infrastructure/nginx/templates-available/sonora.conf.template \
      infrastructure/nginx/templates-active/
   rm infrastructure/nginx/templates-active/bootstrap.conf.template
   docker compose restart nginx
   ```
6. Verify: `curl -I https://$FRONTEND_DOMAIN`, `https://$ADMIN_DOMAIN`,
   `https://$API_DOMAIN/health` should all return real responses over
   HTTPS now.

## Renewal

Certs are valid 90 days. Add a host cron job (outside Docker Compose,
since it just needs to run periodically, not stay running):
```
0 3 * * 1 cd /path/to/sonora && docker compose run --rm certbot renew --webroot -w /var/www/certbot && docker compose restart nginx
```

## What was actually verified in this environment

No real domain exists here, so the HTTPS path itself couldn't be tested
end-to-end (same as Drive/Bandcamp/Dropbox credentials elsewhere). What
WAS verified for real:
- `frontend.Dockerfile` and `admin.Dockerfile` — real production builds
  (`next build`, standalone output), both containers actually started
  and served real pages (`/login`, `/`, `/analytics` all returned 200).
- The `PORT` env var quirk: the standalone server always listens on 3000
  unless `PORT` is set — confirmed by testing (admin container returned
  connection-refused until `PORT=3001` was set), now wired correctly in
  `docker-compose.yml`.
- `nginx -t` on both rendered templates (`envsubst` with test domains) —
  the bootstrap config passed standalone; the full SSL config needed
  dummy self-signed certs at the expected `/etc/letsencrypt/live/...`
  paths plus 3 dummy containers actually named `frontend`/`admin`/`api`
  on a shared Docker network (nginx resolves upstream names at startup,
  so testing it fully isolated fails on DNS, not on the config itself).
  With those in place, `nginx -t` passed clean — the config is
  structurally correct, not just "looks right".
