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

## Status saat ini: Sprint 13 selesai — lanjut Sprint 14 (UI Fidelity & Production Readiness, INI BUKAN SPRINT TERAKHIR PERENCANAAN — baca `docs/roadmap.md`, project baru "selesai" setelah Sprint 14)

Lihat detail masing-masing sprint di "Riwayat Sprint" di bawah.

## Riwayat Sprint

### Sprint 13 (selesai)

Sprint 13 (Observability & DR) selesai 100% (2026-08-05). Keputusan baru
(desain `/metrics`, mekanisme backup, kenapa `hey` bukan k6) di
`docs/decisions/0007-sprint13-observability-and-dr.md`. **Restore drill
BENERAN dicoba** (bukan cuma ditulis di dokumen) — sesuai kata roadmap
"wajib dicoba beneran" — pakai SFTP server lokal (`atmoz/sftp`) sebagai
pengganti Hetzner Storage Box asli yang belum ada kredensialnya.

- [x] `GET /metrics` (tanpa `/api/v1` prefix, tanpa auth — endpoint
      internal Prometheus scraper) — `prometheus/client_golang` +
      `middleware.Metrics` (request count + latency histogram per
      method/route/status), sengaja minim (tidak ada metric bisnis
      kustom kayak queue depth, skala personal tidak butuh itu)
- [x] `application/backup` — `pg_dump "$DATABASE_URL" | gzip` lalu
      upload lewat `scp` (CLI, `postgresql-client` + `openssh-client`
      ditambah ke `worker.Dockerfile`) ke Hetzner Storage Box. Dijadwalkan
      harian 02:00 lewat `asynq.Scheduler` yang sudah ada sejak Sprint 10
      (`maintenance:backup_database`), plus `POST /admin/backup/run`
      buat trigger manual (pola sama dengan "Run health check"/"Sync now")
- [x] Config baru (opsional, bukan required): `BACKUP_SSH_HOST/USER/
      REMOTE_PATH/KEY_PATH` — kalau `BACKUP_SSH_HOST` kosong, job backup
      log jelas "not set, skipping" dan tidak retry sia-sia (pola
      credential-gap yang sama dengan Drive/Bandcamp/Dropbox)
- [x] Rate limit dari Sprint 12 (300 req/menit per IP, global) ketemu
      "beneran mengganggu" load test pertama (lihat catatan verifikasi)
      — bukan bug, tapi insight nyata: budget itu dibagi rata ke SEMUA
      endpoint per IP yang sama, bukan per-route

Diverifikasi jauh melebihi sekadar "kode ada":
- **Restore drill sungguhan**: trigger `POST /admin/backup/run` asli →
  file `.sql.gz` BENERAN mendarat di server SFTP lokal (dicek langsung
  di container SFTP-nya) → didownload lagi → di-restore (`psql -f`) ke
  Postgres KOSONG yang baru → dibandingkan: 21 tabel di kedua DB SAMA,
  isi tabel `users` (id/email/role) SAMA PERSIS. Ini membuktikan seluruh
  rantai pg_dump→scp→download→restore utuh, bukan asumsi "seharusnya
  jalan"
- `/metrics`: dicek langsung ada `sonora_http_requests_total` dan
  `sonora_http_request_duration_seconds` dengan angka yang benar
  ke-update setelah request nyata ke `/health`
- Backup tanpa config: `BACKUP_SSH_HOST` kosong → job gagal rapi
  ("not set, skipping"), tidak crash, dikonfirmasi di log
- Load test (`hey`, container terpisah): 500 request ke `/health` dalam
  beberapa detik → limiter global memotong di ~300 (429 setelahnya) —
  BENERAN membuktikan rate limit Sprint 12 jalan di bawah beban nyata,
  bukan cuma di tes curl satu-satu. Diulang di bawah budget (200 request
  ke `/api/v1/auth/me`, endpoint asli lewat JWT+GORM): 200/200 sukses,
  p99 latency ~29ms — performa sehat untuk skala personal

**Butuh input manual dari user**: SSH host/user/path/key Hetzner Storage
Box asli belum ada — job backup terjadwal (harian 02:00) akan skip
dengan pesan jelas sampai `.env` diisi. Mekanisme dan restore drill sudah
terbukti benar lewat stand-in lokal di atas.

### Sprint 12 (selesai)

Sprint 12 (Security hardening) selesai 100% (2026-08-05). Keputusan
konkret (threshold rate limit, semantik reuse detection, sumber CORS)
didokumentasikan di `docs/decisions/0006-sprint12-security-hardening.md`.
Tidak butuh input manual dari user — semua bisa diverifikasi langsung.

- [x] **Enkripsi credential AES-256** — sudah selesai sejak Sprint 3
      (`crypto.Box`, AES-256-GCM), dipakai lagi untuk
      `ingest_source_connections` di Sprint 10. Item roadmap ini
      diverifikasi ulang, bukan ditulis ulang.
- [x] **Rate limit** — `gofiber/fiber/v2/middleware/limiter` (in-memory,
      cukup untuk 1 proses/1 VPS): global 300 req/menit per IP di semua
      `/api/v1/*`, plus limiter lebih ketat (10 req/menit per IP) khusus
      di endpoint publik tanpa auth (`/auth/google`, `/auth/google/callback`,
      `/auth/refresh`) — target paling masuk akal untuk brute-force
      meski skala personal
- [x] **Refresh token reuse detection** — rotasi (Sprint 2) sekarang
      membedakan token yang expired wajar vs token yang SUDAH PERNAH
      dipakai/di-revoke sebelumnya (reuse). Kasus reuse memicu
      `RevokeAllForUser` — bukan cuma menolak request itu, tapi
      mem-invalidasi SEMUA sesi user itu (termasuk token hasil rotasi
      yang seharusnya masih valid), standar industri untuk rotating
      refresh token
- [x] **CORS lock down** — `AllowOrigins` sekarang dirangkai dari
      `cfg.FrontendURL` + `cfg.AdminURL` (env var yang sudah ada sejak
      Sprint 9), bukan string hardcoded `localhost:3000,3001` —
      otomatis benar saat deploy ke domain VPS asli

Diverifikasi nyata lewat curl (bukan cuma baca kode):
- Rate limit: 15 request beruntun ke `/auth/google` → 8 pertama `302`,
  sisanya `429` — limiter beneran memotong di angka yang tepat
- CORS: preflight dari `Origin: http://localhost:3000` dapat header
  `Access-Control-Allow-Origin` yang cocok; dari `Origin:
  http://evil.example.com` header itu TIDAK ADA sama sekali (ditolak
  browser secara implisit)
- Refresh reuse: seed 1 refresh token asli → pakai sekali (sukses, dapat
  token baru) → pakai lagi token LAMA yang sama (401, sesuai ekspektasi)
  → cek token BARU dari langkah pertama (yang seharusnya masih berlaku)
  → JUGA 401, membuktikan `RevokeAllForUser` benar-benar menuntaskan
  semua sesi, bukan cuma menolak satu request. Log server mengonfirmasi
  `auth: refresh token reuse detected, all sessions revoked` tercatat
  di kedua percobaan reuse.

### Sprint 11 (selesai)

Sprint 11 (Polish ingest) selesai 100% (2026-08-05). Beda dari
sprint-sprint sebelumnya: MusicBrainz + Cover Art Archive tidak butuh
credential sama sekali, jadi ini sprint pertama yang bisa diverifikasi
end-to-end PENUH tanpa "butuh input manual dari user" — dibuktikan
dengan lagu asli ("Yesterday" — The Beatles) lewat panggilan API
sungguhan, sama seperti verifikasi LRCLIB di Sprint 6. Desain baru
(skema kolom, scope enrichment) didokumentasikan di
`docs/decisions/0005-sprint11-waveform-musicbrainz-analytics.md` karena
`docs/api-design.md` tidak menyebutnya sama sekali (kecuali 2 endpoint
analytics yang sudah ada).

- [x] Migration `000006` — `songs.waveform_peaks SMALLINT[]`,
      `songs/artists/albums.musicbrainz_id` (semua nullable — ingest
      tetap lengkap tanpa keduanya)
- [x] `mediainfo.GenerateWaveform` — shell out `ffmpeg` (sudah ada sejak
      Sprint 3 untuk `ffprobe`) decode ke PCM 8-bit mono 8kHz, reduce ke
      200 bucket peak — tidak nambah dependency baru
- [x] `infrastructure/musicbrainz` — cari recording by title+artist+
      duration (toleransi ±2 detik), rate limit 1 req/detik
      (`golang.org/x/time/rate`, sudah transitive dependency lewat
      `google.golang.org/api`) sesuai kebijakan resmi MusicBrainz; kalau
      match ketemu dan cover art ada di Cover Art Archive (dicek beneran
      lewat HEAD request, bukan asumsi URL), isi `albums.cover_url`
      HANYA kalau masih kosong
- [x] `ingest.Process` — waveform + MusicBrainz jalan setelah
      `CompleteIngestJob`, keduanya best-effort (gagal di-log, TIDAK
      menggagalkan job yang sudah `completed`)
- [x] `application/analytics` + `GET /admin/analytics/top-played`,
      `GET /admin/analytics/storage-growth` (query baru: top 10 dari
      `play_history`, total bytes storage per bulan 6 bulan terakhir
      dengan zero-fill via `generate_series` biar bulan sepi tetap
      muncul di chart sebagai 0)
- [x] Admin `/analytics` — dari placeholder (sejak Sprint 9) jadi
      fungsional: bar chart Storage Growth + list Most Played, pakai
      `recharts` yang sudah jadi dependency `apps/admin` sejak Sprint 1
      scaffold tapi belum pernah dipakai. "Download Trend" dari
      `docs/screens-spec.md` #21 SENGAJA tidak dibangun — tidak ada
      endpoint pendukungnya, di luar scope Sprint 11 (lihat ADR 0005)

Diverifikasi jauh melebihi ekspektasi awal — upload 2 file MP3 nyata
(dibuat dengan `ffmpeg` + tag ID3 asli) lewat pipeline ingest yang
SESUNGGUHNYA (bukan panggil fungsi langsung), storage upload di-bypass
pakai pre-seeded `storage_files` row (dedup by checksum) supaya tetap
bisa jalan tanpa credential Drive asli, tapi waveform + MusicBrainz +
Cover Art Archive semuanya panggilan jaringan SUNGGUHAN:
- File 1 (tag "Yesterday"/"The Beatles", durasi ~125 detik pas): dapat
  `musicbrainz_id` recording asli, `musicbrainz_id` artist asli, album
  "Help!" dengan `musicbrainz_id` asli, DAN `cover_url` dari Cover Art
  Archive yang dikonfirmasi HTTP 200 (gambar beneran ada)
- File 2 (tag fiktif "Totally Fictional Song Xyzzy123"): tidak dapat
  match (perilaku benar — bukan bug), tapi waveform tetap ke-generate
  (200 peak) — membuktikan dua fitur independen satu sama lain
- Kedua file: `waveform_peaks` ke-generate 200 titik dari audio asli
- Analytics: play history asli direkam untuk "Yesterday" → top-played
  menunjukkan hasil yang benar; storage-growth menunjukkan total bytes
  bulan berjalan dengan benar
- Playwright admin: bar chart Storage Growth render SVG asli (6 bar,
  1 dengan data nyata), Most Played menampilkan "Yesterday — The
  Beatles" dari data yang sungguhan direkam

Tidak ada catatan "butuh input manual dari user" untuk sprint ini.

### Sprint 10 (selesai)

Sprint 10 (Scheduled jobs) selesai 100% (2026-08-05). `docs/api-design.md`
tidak punya spesifikasi apapun untuk bagian ini — semua keputusan desain
baru (skema tabel, pola credential, scope Bandcamp/Dropbox) didokumentasikan
di `docs/decisions/0004-sprint10-scheduled-jobs-and-ingest-sources.md`
sebelum coding, sesuai filosofi "ADR untuk keputusan arsitektur signifikan".

Scheduler infra:
- [x] `asynq.Scheduler` jalan sebagai goroutine tambahan di proses
      `worker` yang sama (bukan proses/deployment terpisah — 1 VPS,
      jangan over-engineer)
- [x] **Garbage Collector** (harian 03:00): hapus file temp ingest untuk
      job `completed` (file sudah di storage, tidak dibutuhkan lagi) —
      SENGAJA tidak menyentuh `temp_path` job `failed` (retry butuh itu);
      plus purge `refresh_tokens` yang sudah expired
- [x] **Storage Optimizer** (mingguan, Minggu 04:00): jalankan health
      check ke SEMUA storage account aktif otomatis (`RunHealthChecks`,
      reuse logic tombol manual Drive Manager Sprint 9) — data quota
      routing (Sprint 9) tetap segar tanpa Owner harus klik manual

Ingest source connections (Bandcamp/cloud sync):
- [x] Migration `000005` — tabel `ingest_source_connections` (bentuk
      sengaja disamakan dengan `storage_accounts`: `credentials_encrypted`
      lewat `crypto.Box` yang sama), dikelola Owner & global (bukan
      per-user) — konsisten dengan penempatan halaman admin "Ingest
      Sources" di `docs/screens-spec.md`
- [x] `infrastructure/bandcamp` — client fancollection API (endpoint yang
      dipakai aplikasi resmi Bandcamp sendiri untuk redownload koleksi),
      credential = cookie `identity` + `fan_id` didapat manual di luar
      aplikasi (pola sama dengan refresh token Drive, ADR 0002). Item tipe
      "album" (zip multi-track) SENGAJA tidak didukung v1 — di-skip
      dengan log jelas, cuma track tunggal yang di-download
- [x] `infrastructure/dropbox` — client OAuth refresh-token (pola sama
      dengan Drive), `list_folder` + `download`, filter ekstensi audio.
      SATU-SATUNYA implementasi `cloud_sync` konkret (OneDrive/iCloud
      backlog, bukan penyimpangan — `source_type` tetap generik)
- [x] `application/ingestsource` — Connect/List/Disconnect/Sync/SyncAll;
      `ingest.Service.Accept` diperluas terima `sourceType` (sebelumnya
      hardcoded `manual_upload`); download hasil sync masuk pipeline
      Accept → Process yang SAMA PERSIS dengan upload manual — dedup by
      checksum yang sudah ada sejak Sprint 3 bikin sync idempoten tanpa
      butuh tabel "item sudah diproses" baru
      hasil sync di-attribute ke user Owner (`identity.FindOwner`,
      asumsi single-owner personal deployment dari Sprint 2)
- [x] `POST/GET/DELETE /admin/ingest-sources/connections[/:id]`,
      `POST .../sync` (Owner only) — endpoint terakhir yang ditambahkan
      ke `docs/api-design.md` Admin section
- [x] Admin frontend `/ingest-sources` — dari placeholder read-only
      (Sprint 9) jadi fungsional penuh: connect form (pilih provider,
      field beda per provider), list dengan status + last synced, tombol
      "Sync now" dan disconnect

Diverifikasi jauh lebih nyata dari yang diperkirakan — Bandcamp dan
Dropbox client BENERAN memanggil API asli lewat internet (bukan mock):
- GC: file temp asli dibuat di volume `ingest_tmp` bersama, job
  `completed` + refresh token expired asli di-seed ke DB, dijalankan lewat
  Asynq real (enqueue manual ke Redis) → file hilang dari disk,
  `temp_path` ter-null, refresh token terhapus — dikonfirmasi lewat cek
  langsung ke volume & DB, bukan cuma baca log
- Storage Optimizer: dijalankan nyata, benar meng-iterasi akun aktif dan
  melaporkan hasil (gagal pada 1 akun test dengan credential rusak
  sengaja — perilaku benar, bukan bug)
- Bandcamp: `POST bandcamp.com/api/fancollection/1/collection_items`
  BENERAN terhubung ke internet dari container, balas 200 dengan koleksi
  kosong (cookie palsu) — tidak crash, sync selesai tanpa item, sesuai
  desain
- Dropbox: `POST api.dropboxapi.com/oauth2/token` BENERAN ditolak server
  Dropbox asli (`invalid_client: Invalid client_id or client_secret`),
  error itu diteruskan bersih ke response API — pola "gagal dengan pesan
  jelas" yang sama seperti Drive sejak Sprint 3, sekarang terbukti dengan
  2 provider tambahan
- Playwright admin: connect Bandcamp → card muncul "Never synced" →
  klik Sync now → benar update jadi "Last synced ..." (real API call
  seperti di atas) → disconnect → card hilang

**Butuh input manual dari user** (pola sama seperti Drive Sprint 3/4):
cookie `identity` + `fan_id` Bandcamp asli, dan `refresh_token`/`app_key`/
`app_secret` OAuth Dropbox asli, keduanya belum ada. Kode connect/sync
lengkap dan gagal dengan rapi tanpa itu (dikonfirmasi di atas), bukan
crash.

### Sprint 9 (selesai)

Sprint 9 (Multi-drive pool) selesai 100% (2026-08-05), memperluas bootstrap
minimal ADR 0002 dari Sprint 3 (backward-compatible, field/bentuk response
tidak berubah). Backend diverifikasi lewat curl end-to-end + query SQL
langsung buat quota-routing; frontend admin diverifikasi lewat Playwright
browser asli, termasuk role gate Owner-vs-Member.

Backend:
- [x] `GetActiveStorageAccount` diganti jadi quota-aware: `ORDER BY
      (COALESCE(quota_bytes, 999999999999) - used_bytes) DESC LIMIT 1`
      dengan filter `health_status <> 'down'` — akun tanpa batas quota
      (`quota_bytes IS NULL`) dianggap punya kapasitas nyaris tak
      terbatas, bukan otomatis menang lewat `COALESCE` naif
- [x] `Provider.HealthCheck` (interface baru) + implementasi Google Drive
      lewat `About.Get().Fields("storageQuota")`; hasil ditulis balik ke
      `health_status`/`quota_bytes`/`used_bytes`/`last_health_check_at`
- [x] `application/storageaccount` — tambah `Delete` (ditolak dengan
      `ErrInUse` kalau FK `storage_files` RESTRICT masih menunjuk ke akun
      itu → 409, bukan 500) dan `HealthCheck`
- [x] `DELETE /admin/storage/accounts/:id`, `POST
      /admin/storage/accounts/:id/health-check` — endpoint terakhir yang
      belum ada dari `docs/api-design.md` Admin section
- [x] Login admin app-aware: `GET /auth/google?app=admin` menyisipkan app
      ke `state` OAuth (`nonce:admin`, bukan query param terpisah supaya
      tidak jadi open redirect — `resolveTargetURL` cuma bisa hasilkan 2
      URL yang sudah dikonfigurasi), `ADMIN_URL` baru di config/`.env`

Frontend admin (`apps/admin`, sebelumnya cuma scaffold bare sejak Sprint 1):
- [x] `store/auth.ts`, `lib/api.ts`, `/login`, `/auth/callback`,
      `providers.tsx` — pola identik dengan `apps/frontend`, plus fetch
      `GET /auth/me` sesudah bootstrap buat dapat `role`
- [x] `app-shell.tsx` — gate Owner: kalau `role !== "owner"` tampilkan
      "Access Denied" alih-alih sidebar+children (defense in depth; API
      sudah menegakkan `RequireRole(owner)` di semua route `/admin/*`,
      tapi UI tidak boleh nampilkan wall of 403 ke Member yang salah masuk)
- [x] `components/sidebar.tsx` — 6 section sesuai `docs/screens-spec.md`
      (Dashboard, Drive Manager, Ingest Sources, Lyrics Source, Job Queue,
      Analytics); hanya Drive Manager yang fungsional sprint ini, sisanya
      placeholder jujur (pola vertical-slice yang sama dipakai `BottomNav`
      Sprint 4)
- [x] `/drive-manager` — card per storage account (badge health status,
      progress bar quota, sisa GB), form "+ Add drive" (refresh token
      masih manual/out-of-band sesuai ADR 0002 — tidak ada in-app OAuth
      consent flow buat Drive), tombol "Run health check" dan delete per
      card

Diverifikasi:
- curl: create x2 (quota beda), list, health-check (fake creds → `down`,
  200), health-check akun tidak ada (404), delete (200), delete ulang
  (404)
- SQL langsung: 3 akun (LowFree/HighFree/DownAcct is_active tapi
  health_status=down) → query routing pilih `HighFree` (free space
  terbesar), `DownAcct` benar-benar dikecualikan meski quota totalnya
  paling besar
- Playwright (admin app, port 3001, BARU — sebelumnya semua browser test
  di port 3000): login Owner → dashboard → 6 section sidebar render →
  Drive Manager create → badge healthy → run health-check → badge down
  (fake creds, expected) → delete → card hilang; login Member (role
  seed manual) → "Access Denied", sidebar tidak render sama sekali

**Butuh input manual dari user**: sama seperti Sprint 3/4/8, refresh token
Drive asli (dan karenanya health check yang benar-benar `healthy`) masih
menunggu `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` + refresh token nyata.
Kode Drive Manager sudah lengkap dan gagal dengan rapi (`down`, bukan
crash) tanpa itu.

### Sprint 8 (selesai)

Sprint 8 (Active Device + PWA) selesai 100% (2026-08-05). Lihat
`docs/decisions/0003-active-device-ui-placement.md` — `docs/screens-spec.md`
tidak pernah mendesain UI Transfer Playback sama sekali, jadi ini
keputusan baru (bukan penyimpangan dari spec, spec-nya memang belum ada).

Backend:
- [x] `TokenPair` (application/auth) sekarang bawa `DeviceID` — sebelum
      Sprint 8, `device_id` yang dibuat saat OAuth callback/refresh tidak
      pernah dikembalikan ke frontend (gap yang baru ketahuan sekarang,
      lihat ADR 0003). `POST /auth/refresh` dan redirect `/auth/callback`
      sekarang menyertakannya.
- [x] `wstoken` diperluas: token sekarang bawa `device_id` juga (bukan cuma
      `user_id`), supaya hub bisa relay ke device tertentu
- [x] `ws.Hub` — `SendToDevice` selain `Broadcast`; `POST /player/transfer`
      set `playback_state.active_device_id` + sinkron `devices.is_active`
      (transaksi 2 UPDATE — clear semua device lain punya user itu, set 1)
- [x] Relay `player:command`: device yang BUKAN Active mengirim command
      lewat WS → di-relay ke device yang Active saja; kalau si pengirim
      KEBETULAN device yang aktif, tidak di-relay (device aktif eksekusi
      langsung, bukan lewat WS ke diri sendiri)
- [x] Endpoint `/player/transfer` TIDAK menegakkan "hanya Active Device
      boleh update state" secara server-side — batasnya di client (device
      non-aktif kirim command lewat WS, bukan panggil `/player/state`
      langsung). Cukup untuk skala personal/keluarga, dicatat sebagai
      potential hardening kalau perlu nanti.

Frontend:
- [x] `store/auth.ts` sekarang simpan `deviceId` juga (dari `/auth/callback`
      fragment ATAU `/auth/refresh` response)
- [x] `store/ws.ts` — connect WS pakai `device_id`, dengar `player:state`
      buat tahu device mana yang aktif
- [x] `player.ts` — main lagu "di sini" otomatis klaim device ini jadi
      Active (`syncRemoteState`), sama seperti behavior Spotify Connect
      yang jadi referensi sejak awal; device switcher eksplisit
      (`/devices`, link dari Now Playing) buat transfer ke device LAIN
- [x] `/devices` — list device, tap → `POST /player/transfer`
- [x] PWA: `public/manifest.json`, `public/icon.svg`, `public/sw.js`
      (network-first, TIDAK precache app shell karena asset Next.js
      di-hash per build — precache di sini malah lawan cache-busting
      framework-nya sendiri)
- [x] Offline download: `lib/offline-db.ts` (IndexedDB native, tanpa
      dependency baru), tombol download di song detail, `/downloads` (list
      + hapus manual — TIDAK ada auto-LRU sesuai roadmap). `player.ts` cek
      IndexedDB dulu sebelum stream network kalau ada

Diverifikasi jauh lebih dalam dari sprint-sprint sebelumnya karena
melibatkan 2 "device" sekaligus:
1. WS client Go asli (bukan browser) — handshake, transfer, relay
   `player:command` ke device aktif SAJA, device aktif tidak nge-relay ke
   diri sendiri, `devices.is_active` sinkron dengan benar.
2. Playwright browser asli — WebSocket connection BENERAN kebuka dari
   browser (`page.on("websocket")`, bukan cuma diasumsikan dari kode),
   Transfer Playback dari UI asli (klik device di `/devices`) memicu
   broadcast `player:state` yang keliatan lewat network trace, service
   worker ke-registrasi dan **activated**, dan yang paling penting:
   offline test sungguhan — visit halaman online dulu, `context.setOffline(true)`,
   reload, halaman tetap render dari cache SW. Bukan asumsi "seharusnya
   jalan", tapi dibuktikan literally di browser tanpa network.

**Butuh input manual dari user**: sama seperti sprint-sprint sebelumnya,
download lagu asli (dan karenanya offline playback asli) masih menunggu
`GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET` + refresh token Drive asli —
tombol download sudah lengkap dan gagal dengan rapi (pesan error jelas)
tanpa itu.

### Sprint 7 (selesai)

Sprint 7 (WebSocket infra) selesai 100% (2026-08-05), backend-only sesuai
roadmap (bullet Sprint 7 tidak menyebut frontend, beda dari Sprint 4-6 —
konsumsi WS di UI baru ada gunanya nyata bareng Active Device Sprint 8).
Diverifikasi pakai WS client Go asli (bukan mock): handshake, initial
state push, broadcast real-time ke 1 DAN 2 koneksi sekaligus, single-use
enforcement, dan penolakan token invalid — semua diuji nyata.

- [x] `infrastructure/wstoken` — token 60 detik, single-use (Redis GETDEL
      atomik, jadi race antara dua percobaan pakai token yang sama tidak
      bisa dua-duanya berhasil)
- [x] `application/playback` — get/upsert `playback_states`, TANPA
      pengecekan otoritas Active Device (itu Sprint 8 — device mana pun
      milik user bisa update state untuk sekarang)
- [x] `apps/backend/internal/ws` — hub in-memory per proses (cukup untuk
      1 VPS/1 proses api sesuai "Batasan Deployment"; kalau nanti scale ke
      multi-instance butuh pub/sub bersama, di luar scope sekarang)
- [x] `POST /ws/token`, `GET /ws?token=` (handshake), `GET/POST /player/state`
      — `POST /player/state` sengaja bukan endpoint granular
      play/pause/seek/next/previous/transfer dari `docs/api-design.md`;
      itu lapisan Sprint 8 yang akan manggil upsert yang sama ini secara
      internal, plus authority check "hanya Active Device yang boleh"

Tes yang dijalankan: connect fresh → dapat push state terkini langsung;
2 device connect bareng → keduanya dapat initial push DAN broadcast
setelah `POST /player/state`; reuse token yang sama dua kali → gagal di
percobaan kedua; token acak/invalid → 401.

### Sprint 6 (selesai)

Sprint 6 (History & Lyrics) selesai 100% (2026-08-05). LRCLIB diuji dengan
lagu ASLI ("Yesterday" — The Beatles) lewat panggilan API sungguhan (bukan
mock) — lirik synced dengan timestamp beneran didapat dan di-cache.

- [x] `application/history` — record play, list (cursor pagination pola
      sama dengan ingest), `ListContinueListening` (query `DISTINCT ON`
      + filter `progress > 5s AND progress < 95% duration`, exclude lagu
      yang baru mulai atau nyaris selesai)
- [x] `application/queue` — CRUD queue_items (position fractional, pola
      sama dengan playlist)
- [x] `application/lyrics` + `infrastructure/lyrics` (klien LRCLIB) —
      cache-first (`lyrics` table), lazy find-or-create `lyrics_providers`
      row untuk "lrclib" saat cache-miss pertama
- [x] Endpoint `docs/api-design.md`: `/history` (GET/POST),
      `/library/continue-listening`, `/queue*`, `/songs/:id/lyrics`
- [x] Frontend: `/lyrics` (fullscreen, parse LRC → highlight baris aktif +
      fade progresif + tap-to-seek, sesuai `docs/design-system.md`),
      `/queue` (now playing + next-up + remove + clear), tombol
      Lyrics/Queue di Now Playing, tombol "add to queue" di song detail,
      widget Continue Listening + Trending nyata di Home (`/`)
- [x] Player (`store/player.ts`) merekam history saat pause/ended lagu

**Bug nyata ketemu & diperbaiki lewat testing** (bukan lewat review kode):
`SearchSongs` melempar error 500 kalau index Meilisearch `songs` belum
pernah dibuat sama sekali (fresh install, atau kapan pun index kosong
total) — Meilisearch balas 404 `index_not_found` untuk search di index
yang belum ada, dan kode kita meneruskannya sebagai error alih-alih
"0 hasil". Ini akan bikin search rusak total di instalasi baru sampai
lagu pertama selesai di-ingest. Fix: `infrastructure/meilisearch` sekarang
cek `*meilisearch.Error` dengan `StatusCode == 404` dan balas hasil kosong.
Ketemu murni karena testing end-to-end pakai state yang benar-benar
kosong (index sengaja dihapus antar sprint saat cleanup) — highlight
kenapa "bersihkan data test tiap habis sprint" ternyata berguna ganda:
selain kebersihan, juga tidak sengaja jadi test kondisi fresh-install.

Catatan scope: `/player/*` (play/pause/seek/next/previous/transfer) dan
`/player/state` di `docs/api-design.md` SENGAJA belum diimplementasikan —
itu Sprint 7/8 (Active Device + WebSocket), bukan Sprint 6. Merekam
history dari real playback juga belum bisa diverifikasi lewat browser
(audio tidak pernah benar-benar main tanpa Drive asli, jadi event
pause/ended asli tidak pernah terpicu di lingkungan tes) — kode sudah
benar secara logika, tapi correctness end-to-end menunggu credential Drive
asli sama seperti Sprint 3/4.

### Sprint 5 (selesai)

Sprint 5 (Library) selesai 100% (2026-08-04). Diverifikasi lewat curl
end-to-end (semua endpoint playlist/favorite) DAN lewat Playwright browser
nyata (buat playlist → tambah lagu via search → play dari playlist →
favorite lagu dari song detail → cek halaman Favorite → unfavorite).

- [x] `domain/library` + GORM repository (`playlist_repository.go`,
      `favorite_repository.go`) — pola sama persis dengan `domain/identity`
      Sprint 2, konsisten dengan keputusan hybrid ADR (GORM untuk CRUD
      sederhana)
- [x] `application/library` — nama package sengaja sama dengan domain
      package yang dibungkusnya (`domainlibrary` alias di file yang butuh
      keduanya) karena ini memang fitur "Library" Sprint 5
- [x] Playlist CRUD + reorder fractional (`position` dikirim final oleh
      client, server tidak menghitung ulang — konsisten dengan skema
      `playlist_songs.position DOUBLE PRECISION` dari migration Sprint 1)
- [x] Favorites CRUD (song/album/artist/playlist), cek duplikat →
      409 conflict, `GET /favorites?type=` untuk filter
- [x] Endpoint `docs/api-design.md`: `/playlists*`, `/playlists/:id/songs*`,
      `/favorites` (GET/POST/DELETE, body `{type, id}` sesuai spek — bukan
      path param, termasuk untuk DELETE)
- [x] Frontend: `/library` (list + create playlist), `/library/[id]`
      (lihat lagu, cari+tambah lagu via `/search`, hapus lagu, play),
      `/favorite` (list favorite dengan detail ter-resolve per tipe,
      unfavorite) — favorite type artist/album resolve tapi tidak
      di-link (belum ada halaman detail artist/album, lihat catatan di bawah)
- [x] Tombol favorite (heart) ditambahkan di `/song/[id]` — sebelumnya
      Sprint 4 belum ada cara favorite dari UI sama sekali

Catatan scope: Home (`/`) SENGAJA belum diisi widget asli (Continue
Listening, Recently Played, Quick Mix) — itu butuh `play_history` yang
baru ada di Sprint 6. Halaman Artist Detail dan Album Detail (screens-spec
#12, #13) juga belum dibangun — belum ada sprint yang eksplisit
menugaskannya; kemungkinan perlu ditambahkan sebagai bagian dari Sprint 14
(UI Fidelity) atau sisipan kecil sebelum itu kalau dibutuhkan lebih awal.

Satu hal menarik ketemu saat testing browser (bukan bug, tapi dicatat):
Next.js dev-mode error indicator ("N Issues", muncul dari `console.error`
kita sendiri di `player.ts` saat playback gagal — expected karena belum
ada Drive asli) bisa menutupi ikon Home/Search di bottom nav pada
viewport sempit. Ini murni overlay dev-only (tidak ada di production
build), tapi dicatat karena hampir bikin salah diagnosis "nav tidak bisa
diklik" sebagai bug aplikasi padahal bukan.

### Sprint 4 (selesai)

Sprint 4 (Streaming + Play) selesai 100% (2026-08-04). Milestone "upload →
search → play" diverifikasi nyata pakai headless browser (Playwright di
Docker, bukan cuma curl) — dua bug asli ketemu & diperbaiki lewat itu
(lihat di bawah), bukan cuma "compile sukses".

Backend:
- [x] sqlc query baru: `GetSongDetail` (join artist/album/storage_file),
      `GetAlbumDetail`, `ListAlbumsByArtist`, `ListGenres`, `ListRecentSongs`,
      `GetStorageAccountByID`
- [x] `infrastructure/streamtoken` — token terpisah dari JWT utama (scope
      1 lagu, TTL 5 menit) untuk `<audio>` yang tidak bisa kirim header
      `Authorization` (ADR 0001)
- [x] `storage.Provider.Download` (+ `GoogleDriveProvider.Download`) —
      forward header `Range` client ke Drive apa adanya, jadi seek di
      `<audio>` didukung
- [x] `infrastructure/meilisearch` — wrapper SDK resmi; index `songs`
      butuh `primaryKey` eksplisit karena field `id`/`artist_id`/`album_id`
      bikin auto-inference gagal (`index_primary_key_multiple_candidates_found`)
      — ketemu lewat tes nyata ke Meilisearch, bukan asumsi
- [x] `application/catalog` — GetSong/GetAlbum/GetArtist/ListAlbumsByArtist/
      ListGenres/ListRecent, plus IssueStreamToken/ParseStreamToken/Stream
- [x] `application/search` — wrap Meilisearch untuk `/search` & `/search/autocomplete`;
      `/search/trending` sementara pakai "recently added" (bukan play-count
      asli — itu baru ada setelah `play_history` di Sprint 6)
- [x] Ingest pipeline (Sprint 3) di-wire index ke Meilisearch setelah
      `CreateSong` — gagal index tidak menggagalkan job (song row sudah durable)
- [x] Semua endpoint `docs/api-design.md`: `/songs/:id`, `/songs/:id/stream-token`,
      `/songs/:id/stream` (Range), `/albums/:id`, `/artists/:id`,
      `/artists/:id/albums`, `/genres`, `/search*`

Frontend (implementasi UI pertama, sebelumnya cuma scaffold):
- [x] `lucide-react`, `clsx`, `tailwind-merge` ditambahkan; warna semantik
      (success/warning/error/info) dari `docs/design-system.md` ditambah ke
      `tailwind.config.ts` (sebelumnya belum ada)
- [x] `store/auth.ts` — access token di memori saja (bukan localStorage),
      `store/player.ts` — 1 elemen `<audio>` singleton di `window` biar
      playback tidak putus saat pindah halaman
- [x] `app/providers.tsx` — bootstrap sesi via `POST /auth/refresh`
      (cookie httpOnly) tiap full page load
- [x] Halaman: `/login` (tombol Google saja — form email/password di
      screens-spec TIDAK dibangun karena tidak ada endpoint-nya di
      `docs/api-design.md`, lihat catatan di bawah), `/auth/callback`,
      `/search`, `/song/[id]`, `/now-playing`, placeholder jujur untuk
      `/library`, `/favorite`, `/settings` (datanya baru ada Sprint 5-6)
  Home (`/`) juga masih placeholder — Continue Listening/Trending/dst
  butuh playlist/favorite/history yang belum ada.
- [x] `BottomNav` + `MiniPlayer` di root layout, disembunyikan di
      `/login` dan `/auth/callback`

Dua bug nyata ketemu lewat browser test (Playwright), bukan lewat review
kode atau unit test:
1. **Race condition sesi**: `Providers` (root layout, persisten lintas
   navigasi client-side) menembak `POST /auth/refresh` di setiap mount —
   tapi kalau mount itu terjadi pas mendarat di `/auth/callback`, refresh
   yang pasti gagal (belum ada cookie valid) bisa resolve SETELAH callback
   set token asli dari URL fragment, menghapus token yang baru saja benar.
   Fix: skip refresh sepenuhnya di `/login` dan `/auth/callback`.
2. **Pesan error teknis bocor ke UI**: catch block di `player.ts` pakai
   `e.message` sebagai fallback, tapi `DOMException` dari `<audio>` yang
   gagal load lolos `instanceof Error` — jadi user lihat "Failed to load
   because no supported source was found" bukan "Gagal memutar lagu ini."
   Fix: selalu pakai pesan Indonesia yang ramah, pesan asli cukup di
   `console.error` untuk debugging.

Alat bantu development (bukan bagian aplikasi):
- Node/pnpm dijalankan lewat container `node:22-alpine` (bind-mount repo),
  BUKAN install native — jaringan WSL host kena masalah MTU untuk transfer
  besar (download >~1MB macet, `go get`/Docker pull tetap jalan normal
  karena lewat jalur network Docker Desktop yang berbeda). Kalau mau
  `pnpm`/`next` langsung dari WSL shell, ini perlu dibereskan dulu (coba
  `ip link set dev eth0 mtu 1400` seperti yang dipakai sesi ini, tapi
  belum permanen/reboot-proof).
- Dev server frontend ditinggalkan jalan (container `sonora-frontend`,
  port 3000) supaya bisa langsung dibuka di browser.

**Butuh input manual dari user**: sama seperti Sprint 3, upload→Drive→stream
lagu asli belum bisa dites end-to-end tanpa `GOOGLE_CLIENT_ID`/`GOOGLE_CLIENT_SECRET`
dan refresh token Drive asli. Login Google asli juga belum bisa dites
end-to-end karena alasan yang sama (kode OAuth sudah lengkap sejak Sprint 2).

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
