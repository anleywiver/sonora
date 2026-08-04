# Sonora

Personal music streaming platform. Lihat `docs/decisions/` untuk semua
keputusan arsitektur (ADR).

## Struktur

- `apps/frontend` — Next.js 15, user-facing PWA
- `apps/admin` — Next.js 15, admin panel
- `apps/backend` — Go + Fiber, API server
- `apps/worker` — Go + Asynq, background job processor
- `libs/go-core` — shared Go module (domain, application, infrastructure)
- `packages/ui`, `packages/shared-types`, `packages/config` — shared JS/TS

## Setup pertama kali (Sprint 1)

Jalankan semua perintah ini di dalam WSL2 Ubuntu (bukan PowerShell Windows),
lihat Task 0 Environment Setup.

```bash
# 1. Copy env file
cp .env.example .env
# lalu isi GOOGLE_CLIENT_ID, GOOGLE_CLIENT_SECRET, JWT secrets, dst

# 2. Install dependency JS/TS
pnpm install

# 3. Download dependency Go (butuh internet — tidak bisa dilakukan di sandbox)
cd apps/backend && go mod tidy && cd ../..
cd apps/worker && go mod tidy && cd ../..

# 4. Jalankan semua service
docker compose up --build

# 5. Cek health check
curl http://localhost:8080/api/v1/../health
# atau buka http://localhost:8080/health
```

## Development tanpa Docker (lebih cepat untuk iterasi)

```bash
# Terminal 1 — infra saja
docker compose up postgres redis meilisearch

# Terminal 2 — backend
cd apps/backend && go run ./cmd/api

# Terminal 3 — worker
cd apps/worker && go run ./cmd/worker

# Terminal 4 — frontend
pnpm --filter @sonora/frontend dev

# Terminal 5 — admin
pnpm --filter @sonora/admin dev
```

## Catatan penting

Beberapa dependency Go (`fiber`, `asynq`) dan Node belum di-download di
scaffold ini karena dibuat di lingkungan tanpa akses internet. Jalankan
`go mod tidy` (Go) dan `pnpm install` (Node) di mesin kamu sendiri setelah
extract project ini — itu akan mengunduh semua package yang dibutuhkan
sesuai `go.mod` dan `package.json` yang sudah disiapkan.
