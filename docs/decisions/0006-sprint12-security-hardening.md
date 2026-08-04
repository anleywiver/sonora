# ADR 0006 — Security Hardening (Sprint 12)

Status: Diterima
Tanggal: Sprint 12

## Konteks

Roadmap Sprint 12: "Security — rate limit final, enkripsi credential
storage account (AES-256), test refresh token rotation, CORS lock down."
`docs/api-design.md` sudah menyebut `429 rate limited` di konvensi HTTP
status sejak awal, tapi tidak ada middleware rate limit yang benar-benar
terpasang sampai sprint ini. Beberapa keputusan konkret diperlukan.

## Keputusan

### 1. AES-256 credential encryption — sudah selesai sejak Sprint 3, tidak diulang

`infrastructure/crypto.Box` (AES-256-GCM) sudah dipakai untuk
`storage_accounts.credentials_encrypted` sejak Sprint 3 dan
`ingest_source_connections.credentials_encrypted` sejak Sprint 10. Item
roadmap ini sudah terpenuhi — sprint ini hanya memverifikasi ulang
(bukan menulis ulang).

### 2. Rate limit: Fiber `limiter` bawaan (in-memory), bukan Redis-backed

Skala personal/keluarga (1 VPS, bukan produk publik) tidak butuh rate
limiter terdistribusi — store in-memory bawaan `gofiber/fiber/v2/middleware/limiter`
cukup dan tidak menambah dependency baru (satu module yang sama dengan
Fiber). Dua tingkat:
- **Global** (`/api/v1/*`): 300 request/menit per IP — longgar, cuma
  jaring pengaman dari bug client yang nge-loop, bukan pembatas
  pemakaian normal.
- **Auth endpoints** (`/auth/google`, `/auth/refresh`): 10 request/menit
  per IP — endpoint publik tanpa `requireAuth`, target paling masuk akal
  untuk brute-force/credential-stuffing sekalipun skala personal.

### 3. Refresh token rotation: tambah deteksi reuse (bukan cuma rotasi biasa)

Rotasi (revoke lama + issue baru per pemakaian) sudah ada sejak Sprint 2.
Yang ditambah sprint ini: `Refresh` sekarang membedakan dua alasan token
"invalid" — **expired wajar** (tidak butuh aksi lain) vs **REUSE token
yang sudah pernah dipakai/di-revoke sebelumnya** (indikasi refresh token
dicuri dan dipakai lebih dari sekali oleh dua pihak berbeda). Kasus kedua
langsung `RevokeAllForUser` — bukan cuma menolak request itu, tapi
mem-invalidasi SEMUA sesi user itu, termasuk token hasil rotasi yang
"seharusnya" masih valid. Ini standar industri ("refresh token reuse
detection") untuk rotating refresh tokens, bukan penemuan baru — hanya
belum diimplementasikan sejak Sprint 2.

### 4. CORS: origin dari config, bukan string hardcoded

`AllowOrigins` sebelumnya string literal `"http://localhost:3000,http://localhost:3001"`
langsung di `main.go`. Sekarang dirangkai dari `cfg.FrontendURL` +
`cfg.AdminURL` (dua env var yang sudah ada sejak Sprint 9) — supaya saat
deploy ke VPS asli dengan domain asli, CORS otomatis ikut tanpa perlu
edit kode.

## Konsekuensi

- `identity.RefreshTokenRepository` tidak berubah (revoke/reuse logic ada
  di `application/auth`, bukan di repository).
- Reuse detection berarti SATU refresh token lama yang tidak sengaja
  dipakai dua kali (mis. race condition di client — dua tab browser
  refresh bersamaan) akan men-logout SEMUA device user itu, bukan cuma
  gagal sekali. Ini trade-off yang disengaja (keamanan > kenyamanan
  race-condition langka) — dicatat di sini kalau nanti jadi keluhan
  nyata dari pemakaian riil.
- Rate limit in-memory berarti reset kalau proses `api` restart (bukan
  masalah nyata untuk 1 proses/1 VPS) dan tidak konsisten kalau nanti
  `api` di-scale ke >1 instance (di luar scope "Batasan Deployment").
