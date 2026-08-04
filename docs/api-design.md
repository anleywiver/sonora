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
| GET | `/auth/google` | Public |
| GET | `/auth/google/callback` | Public |
| POST | `/auth/refresh` | Refresh token (rotation tiap dipakai) |
| POST | `/auth/logout` | Access token |
| POST | `/auth/logout-all` | Access token, Owner only |
| GET | `/auth/me` | Access token |
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
| GET | `/artists/:id`, `/artists/:id/albums` |
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
| GET/POST | `/admin/jobs`, `/admin/jobs/:id/retry` |
| GET/PATCH | `/admin/lyrics-providers[/:id]` |
| GET | `/admin/analytics/top-played`, `/admin/analytics/storage-growth` |
| GET/POST/DELETE | `/admin/users`, `/admin/users/invite`, `/admin/users/:id` |
