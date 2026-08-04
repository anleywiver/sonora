# ADR 0007 — Observability & Disaster Recovery (Sprint 13)

Status: Diterima
Tanggal: Sprint 13

## Konteks

Roadmap Sprint 13: "Observability & DR — `/metrics` endpoint, automated
backup ke Hetzner box, restore drill wajib dicoba beneran, load test
ringan." `docs/api-design.md` tidak menyebut `/metrics` maupun backup
sama sekali — semua desain di bawah ini baru.

Penting: "Hetzner box" di sini adalah **Hetzner Storage Box untuk backup
database**, BEDA dengan "migrasi storage lagu ke Hetzner" yang disebut di
`CLAUDE.md` (itu untuk file musik kalau quota Drive habis, sengaja
ditunda). Backup database ke object storage terpisah adalah praktik DR
standar, tidak melanggar aturan "jangan migrasi storage lagu duluan".

## Keputusan

### 1. `/metrics`: Prometheus client_golang, metric HTTP dasar saja

`github.com/prometheus/client_golang` (standar de-facto Go) + adaptor
`gofiber/fiber/v2/middleware/adaptor` (sudah satu module dengan Fiber,
bukan dependency baru) untuk expose `promhttp.Handler()` di `/metrics`
(tanpa auth — endpoint metrics internal, expose ke Prometheus scraper
lokal saja, TIDAK expose data sensitif). Metric yang dikumpulkan:
request count + latency histogram per method+route+status (middleware
custom, ~30 baris) plus metric runtime Go bawaan client_golang (goroutine
count, GC, memory). TIDAK menambah metric bisnis kustom (queue depth,
dsb) — skala personal tidak butuh dashboard segitu detail, cukup buat
tahu API masih hidup dan tidak lambat.

### 2. Backup database: `pg_dump` + SFTP, dijadwalkan lewat Scheduler yang sudah ada

`application/backup.Service.RunBackup` shell out `pg_dump
"$DATABASE_URL" | gzip` ke file temp, lalu upload lewat `sftp` (CLI,
`openssh-client` ditambah ke `worker.Dockerfile` bareng
`postgresql-client` untuk `pg_dump`) ke Hetzner Storage Box — otentikasi
SSH key (path di-mount lewat volume, bukan disimpan di DB/kode). Config
baru: `BACKUP_SSH_HOST`, `BACKUP_SSH_USER`, `BACKUP_SSH_REMOTE_PATH`,
`BACKUP_SSH_KEY_PATH`. Dijadwalkan lewat `asynq.Scheduler` yang sudah ada
sejak Sprint 10 (`maintenance:backup_database`, harian 02:00 — sebelum
garbage collector 03:00), bukan infrastruktur baru.

Kenapa shell out ke `pg_dump`/`sftp` alih-alih pure-Go (`pgx` dump
manual, SFTP client library)? `pg_dump` sudah menghasilkan format dump
yang benar dan versi-aware (menangani semua tipe kolom Postgres yang
dipakai project ini termasuk array `SMALLINT[]` dari Sprint 11) — nulis
ulang logic itu di Go murni berisiko salah subtle dan tidak menambah
nilai; CLI Postgres resmi lebih dipercaya untuk tugas ini.

### 3. Restore drill: BENERAN dicoba, pakai stand-in lokal (bukan Hetzner asli)

Kredensial Hetzner Storage Box asli belum ada (pola yang sama dengan
Drive/Bandcamp/Dropbox) — TAPI roadmap eksplisit bilang "restore drill
wajib dicoba beneran", jadi drill ini dijalankan sungguhan terhadap
**SSH/SFTP server lokal pengganti** (container `atmoz/sftp`) sebagai
stand-in Hetzner Storage Box: backup asli dijalankan, file dump di-
download lagi dari "storage box" pengganti itu, di-restore ke database
Postgres KOSONG yang baru, dan hasilnya dibandingkan row-count dengan
database asli. Ini membuktikan mekanisme pg_dump→upload→download→restore
benar-benar utuh, bukan cuma "seharusnya jalan" — cuma target akhirnya
(Hetzner asli) yang menunggu credential.

### 4. Load test ringan: `hey`, bukan k6/Locust

"Ringan" secara eksplisit di roadmap — `hey` (single Go binary, dijalankan
lewat container) cukup untuk beberapa ratus request ke 2-3 endpoint kunci
(`/health`, `/songs/:id`, `/search`). Tidak perlu k6/Locust/Gatling yang
ditujukan untuk skenario beban jauh lebih besar dari yang relevan buat
skala personal/keluarga.

## Konsekuensi

- **Butuh input manual dari user**: SSH host/user/path/key Hetzner
  Storage Box asli belum ada — job backup terjadwal akan gagal dengan
  pesan jelas ("failed to connect") sampai `.env` diisi, sama seperti
  pola Drive/Bandcamp/Dropbox sebelumnya. Mekanisme dan restore drill
  sudah terbukti benar lewat stand-in lokal.
- `/metrics` tidak di-auth — kalau nanti expose API ini ke internet
  publik (bukan cuma 1 VPS internal), tambahkan network-level restriction
  (firewall/reverse-proxy block) alih-alih auth di level aplikasi.
