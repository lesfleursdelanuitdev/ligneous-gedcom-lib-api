# Deploying ligneous-gedcom-lib-api

API behavior and payload examples (including associates `ASSO`/`RELA`) are documented in `API_REFERENCE.md`.

Small HTTP service (parse, validate, enrich, export). Default process port is **8092** in Docker; the Go `-port` flag defaults to **8091** unless `PORT` or `LISTEN_ADDR` overrides it.

## Environment variables

| Variable | Meaning |
|----------|---------|
| `PORT` | TCP port when **`LISTEN_ADDR` is unset** (binds on all interfaces as `:<PORT>`). |
| `LISTEN_ADDR` | Full bind address, e.g. **`127.0.0.1:8092`**. When set, **`PORT` / `-port` are ignored** for the listening socket. Use this behind nginx so the API is not exposed on the public interface. |

## Option A — Docker Compose (simplest)

From **`ligneous-gedcom-lib-api/`** (build context is the parent **`gonsalves-genealogy/`** directory):

```bash
cd /path/to/gonsalves-genealogy/ligneous-gedcom-lib-api
docker compose up -d
```

`restart: unless-stopped` keeps the container up across reboots. Health: `GET http://localhost:8092/health`.

Point **the-gonsalves-family-admin** (and other apps) at the service:

```bash
LIB_API_URL=http://127.0.0.1:8092
# or, if nginx terminates TLS on another host:
LIB_API_URL=https://gedcom-api.yourdomain.com
```

**the-gonsalves-family-admin** loads this from the environment for PM2 (`deployment/ecosystem.config.cjs` defaults to `http://127.0.0.1:8092`). See also `the-gonsalves-family-admin/deployment/README.md` §2 and §8.

## Option B — systemd + nginx (production)

1. Build a static binary on a machine with Go (or in CI), install to `/usr/local/bin/ligneous-gedcom-lib-api`.
2. Copy `deploy/systemd/ligneous-gedcom-lib-api.service` to `/etc/systemd/system/`, adjust `User` / paths if needed.
3. `sudo systemctl daemon-reload && sudo systemctl enable --now ligneous-gedcom-lib-api`
4. Copy `deploy/nginx/ligneous-gedcom-lib-api.conf.example` into your nginx `sites-enabled`, set `server_name` and TLS paths, then `sudo nginx -t && sudo systemctl reload nginx`.

The example unit binds **`127.0.0.1:8092`** so only nginx (on the same host) can reach the Go process.

## Health check

```bash
curl -fsS http://127.0.0.1:8092/health
```

Expect `{"status":"ok"}`.
