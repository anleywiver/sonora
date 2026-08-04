# ADR 0001 — Keputusan Arsitektur Awal Sonora

Status: Diterima
Tanggal: Sprint 1

## Konteks

Dokumen ini memformalkan semua keputusan yang diambil selama fase perencanaan
(STEP 1-11) menjadi catatan tertulis di repository, bukan hanya di riwayat chat.

## Keputusan

### Storage
- Fase awal: Google Drive multi-account sebagai storage pool utama (gratis).
- Trigger migrasi: ketika quota gabungan drive mendekati/terlampaui konsisten.
- Rencana migrasi: tambah Hetzner Storage Box sebagai provider baru di
  `StorageProvider` pool — tidak perlu rewrite karena sudah di belakang interface.

### Repo & Monorepo
- Monorepo (Turborepo + pnpm workspaces untuk JS/TS, `go.work` untuk Go).
- `packages/` khusus JS/TS. `libs/go-core` terpisah untuk Go shared module
  (domain, application, infrastructure) — dipakai bareng oleh `apps/backend`
  dan `apps/worker`.

### Database access
- Hybrid: `sqlc` untuk query performance-critical (search, streaming metadata),
  GORM untuk CRUD sederhana (playlist, favorite, settings, auth).

### Real-time sync
- WebSocket untuk playback state sync antar device.
- Konsep Active Device vs Remote Controller (mirip Spotify Connect) — hanya
  1 device yang benar-benar memutar audio di satu waktu.

### Deployment
- 1 VPS untuk semua service di fase awal (Docker Compose).
- Scale-out multi-server jadi opsi masa depan tanpa perubahan arsitektur
  (karena semua service sudah containerized terpisah sejak awal).

### API
- Pagination: cursor-based untuk semua list endpoint.
- Versioning: URL path (`/api/v1/...`).
- Upload: direct-to-backend dengan streaming + checksum validation sebelum
  relay ke storage (bukan presigned URL, karena Google Drive & Hetzner
  tidak native S3-compatible).
- Endpoint stream & WebSocket pakai short-lived scoped token terpisah dari
  JWT utama (karena `<audio>` element dan WebSocket handshake tidak bisa
  bawa custom Authorization header).

### Ingest
- "Auto Download" direvisi jadi "Auto Ingest" — hanya dari sumber legal
  terautentikasi (manual upload, Bandcamp purchases, cloud storage sync).
  TIDAK ADA scraping dari platform berlisensi (YouTube, Spotify, dll).

### Backup & DR
- Automated daily `pg_dump` ke storage terpisah (Hetzner Storage Box).
- Restore drill wajib dicoba minimal 1x sebelum dianggap "aman" (Sprint 13).

## Konsekuensi

Keputusan-keputusan ini membentuk seluruh struktur kode di Sprint 1 dan
menjadi acuan untuk semua sprint berikutnya. Perubahan terhadap keputusan
di atas harus dicatat sebagai ADR baru, bukan mengubah file ini.
