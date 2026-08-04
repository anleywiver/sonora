# Roadmap — Sonora

Vertical-slice approach: fondasi dulu (1 alur utuh login→upload→dengar),
baru lebar fitur. Jangan loncat sprint.

## PENTING: Sprint 13 BUKAN akhir project

Setelah Sprint 13, sistem functionally jalan tapi UI belum tentu 100%
cocok dengan mockup desain, dan belum production-ready secara deployment.
Lanjut ke **Sprint 14** di bawah sebelum menyatakan project "selesai".

## Fase 6 — UI Fidelity & Production Readiness
- **Sprint 14**: Cocokkan SEMUA 21 halaman persis dengan `docs/screens-spec.md`
  dan `docs/design-system.md` (jangan improvisasi styling). Rename section
  admin "Crawler" jadi "Ingest Sources" (koreksi konsep lama). Setup deployment
  produksi: Nginx + SSL (Let's Encrypt), domain, environment production
  terpisah dari dev. Review checklist final terhadap semua Functional
  Requirement di STEP 1 (lihat ADR) — pastikan tidak ada yang kelewat.

## Fase 0 — Foundation
**Sprint 1**: Monorepo, Docker Compose, migration, CI skeleton.
DoD: `docker compose up` semua service healthy, `/health` 200. **← STATUS SAAT INI**

## Fase 1 — Core Loop (MVP)
- **Sprint 2**: Auth — Google OAuth, JWT + refresh rotation, `devices` table.
- **Sprint 3**: Ingest dasar — upload endpoint, checksum dedup, upload ke 1 akun Google Drive, extract metadata dasar (ffprobe).
- **Sprint 4**: Streaming + Play — stream endpoint (Range + stream-token), HTML5 Audio player, mini player, search via Meilisearch.
  → Milestone: upload lagu → search → play, benar-benar kedengaran.

## Fase 2 — Personalization
- **Sprint 5**: Library — Playlist CRUD + reorder (fractional position), Favorite, Home page data asli.
- **Sprint 6**: History & Lyrics — History, Continue Listening, Queue, Lyrics (LRCLIB) + fullscreen auto-scroll.

## Fase 3 — Multi-Device
- **Sprint 7**: WebSocket infra — handshake ws-token, broadcast `playback_state` dasar.
- **Sprint 8**: Active Device + PWA — Transfer Playback, PWA manifest + service worker, offline download (manual eviction only, TIDAK ada auto-LRU).

## Fase 4 — Automation
- **Sprint 9**: Multi-drive pool — tambah akun Drive, quota-aware routing, health check, Drive Manager admin.
- **Sprint 10**: Scheduled jobs — Asynq Scheduler (daily/weekly update, garbage collector, storage optimizer), provider Bandcamp/cloud sync.
- **Sprint 11**: Polish ingest — waveform generation, metadata MusicBrainz, admin Analytics.

## Fase 5 — Hardening
- **Sprint 12**: Security — rate limit final, enkripsi credential storage account (AES-256), test refresh token rotation, CORS lock down.
- **Sprint 13**: Observability & DR — `/metrics` endpoint, automated backup ke Hetzner box, **restore drill wajib dicoba beneran**, load test ringan.

## Backlog (sengaja ditunda)

- Migrasi Hetzner Storage Box — HANYA kalau quota Google Drive gabungan sudah konsisten terlampaui, bukan preventif
- Crossfade & Gapless playback
- Provider metadata/lyrics tambahan di luar MusicBrainz + LRCLIB
- Notification system

## Task Breakdown — Sprint 1 (referensi, sudah hampir selesai)

Lihat checklist di `CLAUDE.md` bagian "Status saat ini" untuk progress terkini.

## Prinsip Task Breakdown untuk Sprint Selanjutnya

Jangan breakdown semua sprint di depan. Begitu mulai sprint baru:
1. Baca goal & deliverable sprint itu dari roadmap di atas
2. Breakdown jadi task granular (referensi gaya breakdown Sprint 1 di git history / ADR)
3. Kerjakan, commit per task supaya progress terlihat jelas
4. Setelah sprint selesai, update checklist status di `CLAUDE.md`
