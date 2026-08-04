# ADR 0002 — Minimal Storage Account Bootstrap di Sprint 3

Status: Diterima
Tanggal: Sprint 3

## Konteks

Roadmap Sprint 3 ("Ingest dasar") mensyaratkan upload lagu benar-benar
tersimpan ke 1 akun Google Drive. Tapi seluruh endpoint manajemen storage
account (`/admin/storage/accounts*`, quota-aware routing, health check,
Drive Manager UI) baru dijadwalkan di Sprint 9 ("Multi-drive pool").

Tanpa endpoint apapun untuk mendaftarkan storage account, pipeline ingest
Sprint 3 tidak punya cara memilih akun Drive tujuan — bahkan secara prinsip,
di luar soal credential asli yang belum tersedia.

## Keputusan

Sprint 3 menambahkan endpoint admin minimal, HANYA create + list:

- `POST /admin/storage/accounts` (Owner only) — daftarkan 1 storage account
  (`label`, `account_email`, `refresh_token` mentah dari OAuth consent Drive
  yang dilakukan manual di luar aplikasi untuk saat ini). Refresh token
  dienkripsi (AES-256-GCM, key dari `STORAGE_CREDENTIALS_ENCRYPTION_KEY`)
  sebelum disimpan ke `credentials_encrypted`.
- `GET /admin/storage/accounts` (Owner only) — list ringkas.

Ingest service memilih storage account aktif pertama (`is_active = true`,
`ORDER BY created_at ASC LIMIT 1`) — tidak ada quota-aware routing.

**TIDAK termasuk** di Sprint 3 (tetap Sprint 9): health-check endpoint,
quota tracking otomatis, multi-account routing, Drive Manager admin UI,
`DELETE`/health-check endpoint.

## Konsekuensi

- Endpoint `/admin/storage/accounts` yang dibangun di Sprint 3 ini akan
  diperluas (bukan diganti) saat Sprint 9 — field dan bentuk response
  dijaga backward-compatible.
- Karena `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` asli belum ada di `.env`,
  dan refresh token Drive juga harus diperoleh manual di luar aplikasi,
  fitur upload-ke-Drive ini TIDAK bisa diuji end-to-end sampai user mengisi
  credential asli. Kode pipeline (checksum dedup, ffprobe, job tracking)
  tetap diverifikasi lewat build + smoke test tanpa storage account aktif
  (job harus gagal dengan pesan jelas "no active storage account", bukan
  crash).
