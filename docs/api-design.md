# API Design — Sonora

Base URL: `https://api.sonora.local/api/v1` (dev) — semua endpoint di bawah relatif ke ini.

## Konvensi

- Auth: `Authorization: Bearer <access_token>` (JWT, expire 15 menit)
- Pagination: `?cursor=<opaque>&limit=20` → `{ data: [...], next_cursor, has_more }`
- Response sukses: `{ "data": {...} }`
- Response error: `{ "error": { "code", "message", "request_id" } }`
- Idempotency: header `Idempotency-Key` didukung di endpoint yang create resource (`/ingest/upload`, `/playlists`, `/history`)

## Auth & Session

| Method | Endpoint | Auth |
|---|---|---|
| GET | `/auth/config` | Public — Sprint 14 sisipan, ADR 0012. `{google_oauth_enabled, app_name}`, dibaca dari `app_settings` |
| GET | `/auth/google` | Public — 403 kalau `app_settings.google_oauth_enabled=false` |
| GET | `/auth/google/callback` | Public — 403 kalau `app_settings.google_oauth_enabled=false` |
| POST | `/auth/login` | Public, rate-limited 5/menit/IP — Sprint 14 sisipan, ADR 0012. Body `{username, password}`, role member |
| POST | `/auth/login/admin` | Public, rate-limited 5/menit/IP — Sprint 14 sisipan, ADR 0012. Body `{email, password}`, HARUS role owner |
| POST | `/auth/refresh` | Refresh token (rotation tiap dipakai) |
| POST | `/auth/logout` | Access token |
| POST | `/auth/logout-all` | Access token, Owner only |
| GET | `/auth/me` | Access token |
| PUT | `/auth/me` | Access token — Sprint 14 sisipan, ADR 0009. Body `{name?, avatar_url?}`; `avatar_url` harus `data:image/...` (thumbnail kecil, bukan link Drive — lihat ADR) |
| GET | `/devices` | Access token |
| DELETE | `/devices/:id` | Access token |

## Catalog

| Method | Endpoint |
|---|---|
| GET | `/songs/:id` |
| POST | `/songs/:id/stream-token` — return short-lived token (5 menit), dipakai di `<audio src>` |
| GET | `/songs/:id/stream?token=` — proxy stream, support `Range` header |
| GET | `/songs/:id/lyrics` |
| GET | `/albums/:id` |
| GET | `/artists/:id`, `/artists/:id/albums`, `/artists/:id/songs` (Sprint 14) |
| GET | `/genres` |

## Search

| Method | Endpoint |
|---|---|
| GET | `/search?q=&type=` |
| GET | `/search/autocomplete?q=` |
| GET | `/search/trending` |

## Library

| Method | Endpoint |
|---|---|
| GET/POST | `/playlists` |
| GET/PATCH/DELETE | `/playlists/:id` |
| POST/PATCH/DELETE | `/playlists/:id/songs[/:song_row_id]` — position pakai fractional (float) |
| GET/POST/DELETE | `/favorites` — `{ type: song\|album\|artist\|playlist, id }` |
| GET/POST | `/history` |
| GET | `/library/continue-listening` |
| GET | `/library/songs`, `/library/albums`, `/library/artists` (Sprint 14 sisipan, ADR 0011 — `?search=`, `?sort=alpha\|recent`, seluruh katalog bukan cuma favorit) |

## Player & Queue

| Method | Endpoint |
|---|---|
| GET | `/player/state` |
| POST | `/player/play`, `/player/pause`, `/player/seek`, `/player/next`, `/player/previous` |
| POST | `/player/transfer` — `{ device_id }`, Active Device handoff |
| GET/POST | `/queue` |
| PATCH/DELETE | `/queue/:id` |
| DELETE | `/queue` — clear all |

## Ingest

| Method | Endpoint |
|---|---|
| POST | `/ingest/upload` — multipart streaming, checksum dedup |
| GET | `/ingest/jobs?status=` |
| GET | `/ingest/jobs/:id` |
| POST | `/ingest/jobs/:id/retry` |
| DELETE | `/ingest/jobs/:id` |

## WebSocket

| Channel | Arah | Catatan |
|---|---|---|
| `ws://.../ws?token=` | Connect | Token dari `POST /ws/token` (short-lived, 60 detik, sekali pakai) |
| `player:state` | Server→Client | Broadcast playback state |
| `player:command` | Client→Server | Command dari controller device |
| `ingest:progress` | Server→Client (admin) | Progress job real-time |
| `drive:health` | Server→Client (admin) | Health check status |

## Admin (Owner only)

| Method | Endpoint |
|---|---|
| GET | `/admin/dashboard` |
| GET/POST/DELETE | `/admin/storage/accounts[/:id]` |
| POST | `/admin/storage/accounts/:id/health-check` |
| GET/POST/DELETE | `/admin/ingest-sources/connections[/:id]` (Sprint 10, ADR 0004) |
| POST | `/admin/ingest-sources/connections/:id/sync` |
| GET/POST/DELETE | `/admin/ingest-sources/:source_type/filters[/:id]` (Sprint 14 sisipan, ADR 0008 — `source_type` is `bandcamp` or `cloud_sync` only, never `manual_upload`) |
| POST | `/admin/backup/run` (Sprint 13, ADR 0007) |
| GET | `/metrics` (Sprint 13 — no auth, no `/api/v1` prefix, ADR 0007) |
| GET/POST | `/admin/jobs`, `/admin/jobs/:id/retry` |
| GET/PATCH | `/admin/lyrics-providers[/:id]` |
| GET | `/admin/analytics/top-played`, `/admin/analytics/storage-growth` |
| GET | `/admin/users` |
| POST | `/admin/users/invite` (Sprint 14 sisipan, ADR 0009 — tidak kirim email, lihat catatan) |
| POST | `/admin/users` (Sprint 14 sisipan, ADR 0012 — "Add User" kredensial: `{username, name?, password, email?}`, aktif langsung, tanpa klaim) |
| DELETE | `/admin/users/:id` (403 kalau target Owner) |
| GET/PATCH | `/admin/settings` (Sprint 14 sisipan, ADR 0012 — key-value: `google_oauth_enabled`, `maintenance_mode`, `app_name`, `default_language`. PATCH body `{key, value}`) |
