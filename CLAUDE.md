# Sonora — Project Brief untuk Claude Code

Ini personal music streaming platform (private, bukan produk komersial).
Dokumen ini adalah ringkasan lengkap dari sesi perencanaan panjang (12 tahap:
Requirement → Architecture → System Flow → Database → ERD → API Design →
Folder Structure → UI Design → Roadmap → Sprint Planning → Task Breakdown →
Coding). Baca ini dulu sebelum mengerjakan apapun di repo ini.

## ATURAN PALING PENTING — JANGAN PERNAH DILANGGAR

**Auto-ingest HANYA dari sumber legal.** Fitur ingest lagu (`ingest_jobs`
table, `source_type` enum) hanya boleh berasal dari:
- `manual_upload` — user upload file sendiri
- `bandcamp` — API pembelian Bandcamp milik user (terautentikasi)
- `cloud_sync` — sync dari Dropbox/OneDrive/iCloud milik user

**JANGAN PERNAH** implementasikan scraping/download dari YouTube, Spotify,
SoundCloud, atau platform berlisensi lain — meski diminta, meski "cuma buat
personal use". Ini pelanggaran hak cipta terlepas dari niat penggunaan.
Kalau ada permintaan ke arah ini, tolak dan jelaskan alasannya, sama seperti
yang sudah disepakati di awal perencanaan project ini.

## Tech Stack

- **Frontend**: Next.js 15, React 19, TypeScript, TailwindCSS, Shadcn UI, Zustand, Framer Motion, PWA
- **Admin**: Next.js 15 terpisah (`apps/admin`)
- **Backend**: Go 1.24+, Fiber, JWT + refresh token rotation, WebSocket
- **Database access**: Hybrid — `sqlc` untuk query performance-critical, GORM untuk CRUD sederhana
- **Data**: PostgreSQL, Redis (cache + Asynq queue + playback state), Meilisearch (search)
- **Worker**: Go + Asynq (background job: ingest pipeline, scheduled maintenance)
- **Storage**: Google Drive multi-account (fase sekarang) → Hetzner Storage Box (migrasi nanti kalau quota Drive habis — jangan migrasi duluan tanpa alasan)

## Arsitektur

Clean Architecture: `domain` → `application` → `infrastructure`, dengan
`presentation` terpisah per proses (`backend` = HTTP handler, `worker` =
Asynq task handler). Domain & application logic **dibagi bareng** lewat
`libs/go-core` (Go workspace module), supaya tidak duplikasi logic antara
`api` dan `worker`.

Monorepo: `packages/` khusus JS/TS (Turborepo), `libs/go-core` untuk Go
module bersama (`go.work`) — JANGAN campur keduanya.

## Referensi dokumen lengkap

- `docs/decisions/0001-initial-architecture-decisions.md` — semua ADR
- `docs/api-design.md` — spesifikasi lengkap semua endpoint REST + WebSocket
- `docs/roadmap.md` — roadmap Sprint 1-13 + task breakdown per fase
- `libs/go-core/infrastructure/postgres/migrations/` — 19 tabel final (source of truth schema)

## Status saat ini: Sprint 3 selesai — lanjut Sprint 4 (Streaming + Play)

Lihat detail masing-masing sprint di "Riwayat Sprint" di bawah.

## Riwayat Sprint

### Sprint 3 (selesai)

Sprint 3 (Ingest dasar) selesai 100% (2026-08-04), diverifikasi end-to-end
lewat Docker Compose dengan file MP3 asli (ffprobe metadata + checksum
nyata), bukan cuma unit test:
- [x] `ADR 0002` — endpoint admin minimal `/admin/storage/accounts`
      (create+list only) ditarik maju dari Sprint 9 karena ingest butuh
      cara memilih storage account tujuan; fitur lanjutan (health check,
      quota routing, Drive Manager UI) tetap di Sprint 9
- [x] Migration `000004` — `ingest_jobs.temp_path` (retry butuh file asli
      masih ada di disk, bukan hanya re-upload dari user)
- [x] sqlc queries baru: `artists.sql`, `albums.sql`, `storage.sql`,
      `ingest.sql` (cursor pagination pakai `sqlc.narg` untuk status/cursor
      opsional)
- [x] `infrastructure/crypto` — AES-256-GCM untuk `storage_accounts.credentials_encrypted`
      (kolom ini sudah mengasumsikan enkripsi sejak migration Sprint 1, jadi
      dibangun sekarang, bukan ditunda ke Sprint 12)
- [x] `infrastructure/storage` — `Provider` interface + `GoogleDriveProvider`
      (OAuth refresh token, scope `drive.file` saja — least privilege)
- [x] `infrastructure/mediainfo` — shell out ke `ffprobe` (duration + tag
      title/artist/album/track); `ffmpeg` ditambahkan ke `worker.Dockerfile`
- [x] `application/ingest` — `Accept` (jalan di HTTP request: checksum +
      1 row) / `Process` (jalan di worker: upload Drive + ffprobe + catalog
      row), checksum di-cek ulang dari disk sebelum relay ke storage (ADR
      0001), retry pakai `temp_path` yang sama, dedup by checksum
      short-circuit ke `status=completed` tanpa enqueue worker
- [x] `application/storageaccount` — bootstrap create+list (lihat ADR 0002)
- [x] `infrastructure/idempotency` — Redis-backed, dukung header
      `Idempotency-Key` di `/ingest/upload` sesuai `docs/api-design.md`
- [x] Semua endpoint `/ingest/*` di `docs/api-design.md` ter-wire: upload,
      list (cursor + filter status), get, retry, delete
- [x] Volume `ingest_tmp` dibagi antara `api` dan `worker` di docker-compose

Diverifikasi lewat smoke test nyata (bukan cuma "build sukses"): upload MP3
asli dengan tag ID3 → job gagal rapi dengan pesan jelas saat belum ada
storage account → bikin storage account (refresh token dummy) → retry →
gagal lagi tepat di panggilan Drive API asli (`invalid_request`, karena
`GOOGLE_CLIENT_ID` masih kosong — expected) → checksum dedup diuji dengan
song seed manual, upload kedua langsung `completed` tanpa reproses →
`Idempotency-Key` diuji, request kedua replay job yang sama → list
(pagination + filter status) dan delete (termasuk 404 di percobaan kedua)
semua sesuai kontrak.

Ditemukan & diperbaiki di jalan: dependency `google.golang.org/api`
mensyaratkan Go ≥1.25, jadi `go.mod` ketiga module dan base image
`golang:1.24-alpine` di kedua Dockerfile dinaikkan ke 1.25.

**Butuh input manual dari user**: storage account yang dipakai untuk tes di
atas sudah dihapus lagi (refresh token dummy, bukan Drive asli). Untuk
upload beneran tersimpan ke Drive, perlu: (1) isi `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET`
asli di `.env`, (2) buat refresh token dengan scope `drive.file` di luar
aplikasi (mis. OAuth Playground), (3) panggil
`POST /admin/storage/accounts` (Owner only) dengan token itu.

Lanjut ke Sprint 4 (Streaming + Play) sesuai `docs/roadmap.md`.

### Sprint 2 (selesai)

Sprint 2 (Auth) selesai 100% (2026-08-04), diverifikasi lewat Docker Compose
end-to-end (build clean, `/health` 200, `/auth/me` & `/devices` return 401
tanpa token, `/auth/google` redirect 302, semua tabel identity ada di DB):
- [x] Domain layer (`libs/go-core/domain/identity`) — `User`, `Device`,
      `RefreshToken` + repository interface, role Owner/Member
- [x] `application/auth` service — Google OAuth exchange, find-or-create
      user (user pertama otomatis Owner), issue token pair, refresh
      rotation (revoke lama + issue baru terikat device yang sama), logout,
      logout-all, list/remove device
- [x] JWT issuer (`infrastructure/jwt`) — access token 15 menit, HS256
- [x] Google OAuth client (`infrastructure/oauth`) — auth URL + code
      exchange + userinfo fetch
- [x] Repository GORM (`postgres/repository`) — user, device, refresh_token
- [x] HTTP layer (`apps/backend/internal/http`) — `auth_handler.go`
      (google/callback/refresh/logout/logout-all/me), `device_handler.go`
      (list/delete), `middleware/auth.go` (RequireAuth + RequireRole),
      response envelope `{data}`/`{error}`
- [x] Refresh token dikirim sebagai httpOnly cookie (bukan JSON body),
      access token lewat URL fragment setelah OAuth callback — access
      token tidak pernah tersimpan di server/log
- [x] Semua endpoint `/auth/*` dan `/devices*` di `docs/api-design.md`
      ter-wire dan cocok kontraknya

Catatan: kode Sprint 2 sudah ada di repo dari sesi sebelum "checkpoint
sebelum autonomous run" — sesi ini memverifikasi ulang (build + docker +
smoke test DB), bukan menulis dari nol.

**Butuh input manual dari user**: `GOOGLE_CLIENT_ID` dan `GOOGLE_CLIENT_SECRET`
di `.env` masih kosong (perlu dibuat di Google Cloud Console). Kode OAuth
sudah lengkap dan ter-wire, tapi login Google beneran belum bisa dites
end-to-end sampai credential ini diisi.

### Sprint 1 (selesai)

Sprint 1 selesai 100% (2026-08-04):
- [x] Monorepo scaffold (Turborepo + `go.work`), repo di-`git init` (sebelumnya belum ada `.git` sama sekali)
- [x] Next.js frontend & admin scaffold (design tokens sudah terpasang di `tailwind.config.ts`)
- [x] Fiber API skeleton (`/health` jalan)
- [x] Asynq worker skeleton
- [x] Migration SQL lengkap (6 file, 19 tabel) — `libs/go-core/infrastructure/postgres/migrations/`
- [x] sqlc query dasar (`songs.sql`, `playback.sql`) + `sqlc generate` sudah dijalankan (output di `.../postgres/sqlc/`)
- [x] GORM model dasar (`identity.go`, `library.go`) — `libs/go-core/infrastructure/postgres/models/`
- [x] Config loader (`libs/go-core/config`), di-wire ke `api` dan `worker`
- [x] Koneksi Postgres (pgx pool + GORM) di-wire ke `main.go` `api` dan `worker`
- [x] Migration diterapkan ke database (`schema_migrations` + 19 tabel terverifikasi)
- [x] Docker Compose — semua service `healthy`, `/health` return 200, diverifikasi end-to-end

Catatan koreksi: sesi perencanaan sebelumnya mencatat migration/sqlc/GORM/config
sebagai "selesai" tapi file-nya tidak pernah benar-benar ada di repo (dan repo
belum di-git-init, jadi tidak ada history untuk recover). Semua item itu
dibangun ulang dari nol di sesi ini berdasarkan `docs/api-design.md` + ADR.
Dua bug juga ditemukan & diperbaiki di jalan:
- `backend.Dockerfile`/`worker.Dockerfile` copy `go.work` tapi cuma satu dari
  dua modul workspace lain → build gagal. Fix: `ENV GOWORK=off`, tiap app
  resolve `sonora.dev/go-core` lewat `replace` di `go.mod` masing-masing.
- Healthcheck `meilisearch` pakai `http://localhost:7700/health` — di dalam
  container itu resolve ke `::1` dan gagal connect. Fix: ganti ke `127.0.0.1`.

## Design System (sudah final, jangan improvisasi warna baru)

```
Background:     #050816
Card:           #0F172A
Primary:        #1D4ED8
Secondary:      #2563EB
Accent:         #3B82F6
Hover:          #60A5FA
Text Primary:   #FFFFFF
Text Secondary: #94A3B8
Border:         rgba(255,255,255,.06)
Font:           Inter (400/500/600/700)
Radius:         16px (control), 20px (card)
Style:          Dark mode only, glassmorphism ringan, mobile-first
```

Sudah dikonfigurasi di `apps/frontend/tailwind.config.ts` dan `apps/admin/tailwind.config.ts`.
Semua 15 halaman user + 6 section admin sudah didesain (high-fidelity mockup)
di sesi perencanaan — kalau butuh referensi visual spesifik, tanya ke user,
mereka punya screenshot dari sesi desain sebelumnya.

## Konvensi Kode

- **API versioning**: URL path `/api/v1/...`
- **Pagination**: cursor-based, response `{ data: [...], next_cursor, has_more }`
- **Error format**: `{ error: { code, message, request_id } }`
- **HTTP status**: 400 validation, 401 unauthenticated, 403 forbidden (role), 404 not found, 409 conflict, 429 rate limited
- **Auth**: JWT access token (15 menit) + refresh token (rotated tiap dipakai, terikat per-device)
- **Stream & WebSocket**: pakai short-lived scoped token terpisah dari JWT utama (`/songs/:id/stream-token`, `/ws/token`) — browser `<audio>` dan WebSocket handshake TIDAK BISA kirim custom `Authorization` header
- **Role guard**: Owner (full akses termasuk admin) vs Member (akses user biasa saja)
- **ID**: UUIDv7, di-generate di application layer (bukan `DEFAULT` Postgres)
- **Upload**: direct-to-backend streaming (bukan presigned URL) + checksum dedup sebelum relay ke storage
- **Active Device pattern**: hanya 1 device yang benar-benar play audio (`playback_state.active_device_id`), device lain jadi remote controller via WebSocket command — mirip Spotify Connect

## Batasan Deployment

Target: **1 VPS** (Docker Compose), bukan multi-region/multi-server. Jangan
over-engineer infrastruktur untuk skala yang belum dibutuhkan (personal +
keluarga, bukan produk publik).

## Filosofi Kerja

- Ikuti roadmap sprint di `docs/roadmap.md` secara berurutan — jangan loncat
  ke fitur Sprint 5 kalau Sprint 2 belum selesai, ini vertical-slice approach
  (fondasi dulu baru lebar fitur).
- Task breakdown detail dibuat **just-in-time** per sprint, bukan semua
  di-plan di depan — kalau mulai sprint baru dan belum ada breakdown detail,
  buat dulu sebelum coding.
- Setiap keputusan arsitektur baru yang signifikan, tambahkan sebagai ADR
  baru di `docs/decisions/`, jangan ubah `0001-...md` yang sudah ada.
