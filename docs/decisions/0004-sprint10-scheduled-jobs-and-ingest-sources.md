# ADR 0004 — Scheduled Jobs & Ingest Source Connections (Sprint 10)

Status: Diterima
Tanggal: Sprint 10

## Konteks

Roadmap Sprint 10: "Scheduled jobs — Asynq Scheduler (daily/weekly update,
garbage collector, storage optimizer), provider Bandcamp/cloud sync."
`docs/api-design.md` tidak punya spesifikasi endpoint atau skema apapun
untuk bagian ini — beda dari sprint-sprint sebelumnya yang selalu punya
kontrak endpoint eksplisit untuk diikuti. Beberapa keputusan desain baru
diperlukan sebelum coding.

## Keputusan

### 1. Scheduler hidup di proses `worker` yang sama, bukan proses terpisah

`asynq.Scheduler` dijalankan sebagai goroutine tambahan di
`cmd/worker/main.go`, bukan binary/deployment terpisah. Konsisten dengan
"Batasan Deployment: 1 VPS, jangan over-engineer" — tidak ada alasan kuat
untuk proses ketiga hanya demi cron.

### 2. Dua job periodik: Garbage Collector (harian) + Storage Optimizer (mingguan)

Frasa roadmap "daily/weekly update, garbage collector, storage optimizer"
dibaca sebagai: scheduler menjalankan job pada kadensi harian/mingguan,
dan jobnya ada dua (garbage collector, storage optimizer) — bukan tiga
item terpisah.

- **Garbage Collector** (harian, 03:00): hapus file temp ingest untuk job
  yang **sudah `completed`** (file sudah ter-upload ke storage, tidak
  dibutuhkan lagi) — SENGAJA TIDAK menyentuh `temp_path` milik job
  `failed`, karena `RetryJob` butuh file itu tetap ada untuk retry.
  Menghapusnya di sini akan diam-diam mematahkan fitur retry dari Sprint 3.
  Juga menghapus baris `refresh_tokens` yang sudah `expires_at < now()`
  (housekeeping ringan, bukan bagian dari security hardening Sprint 12).
- **Storage Optimizer** (mingguan, Minggu 04:00): jalankan health check
  (`storageaccount.Service.HealthCheck`, yang sama dipakai tombol manual
  Drive Manager Sprint 9) ke SEMUA storage account aktif secara otomatis.
  Ini membuat data quota-aware routing (Sprint 9) tetap segar tanpa Owner
  harus klik "Run health check" manual tiap akun.

### 3. Ingest source connections: tabel baru, dikelola Owner (bukan per-user)

`docs/screens-spec.md` #18 menempatkan "Ingest Sources" sebagai halaman
**admin**, sama seperti Drive Manager — bukan halaman self-service tiap
user biasa. Untuk skala personal/keluarga (bukan produk komersial), 1
koneksi Bandcamp + 1 koneksi cloud sync yang dikelola Owner sudah cukup,
konsisten dengan pola `storage_accounts` (juga global, bukan per-user).

Tabel baru `ingest_source_connections` (migration 000005): bentuknya
sengaja disamakan dengan `storage_accounts` (`credentials_encrypted`
lewat `crypto.Box` yang sama, `last_synced_at` alih-alih
`last_health_check_at`) — pola yang sudah terbukti dari Sprint 3/9,
jangan buat pola baru untuk masalah yang sama.

Job hasil sync (bandcamp/cloud_sync) di-attribute ke user Owner (asumsi
single-owner personal deployment dari Sprint 2 — "user pertama otomatis
Owner"). `identity.UserRepository` dapat method baru `FindOwner`.

### 4. Bandcamp: fan-collection API dengan `identity` cookie manual, bukan OAuth

Bandcamp tidak punya OAuth publik untuk fan (pembeli) mengakses koleksi
pembelian sendiri secara terprogram — beda dari Google/Dropbox. API yang
tersedia (`bandcamp.com/api/fancollection/1/collection_items`) hanya bisa
diakses dengan cookie sesi `identity` dari akun yang sudah login, dan itu
API resmi yang dipakai aplikasi Bandcamp sendiri untuk fitur "download
koleksi" — bukan scraping/reverse-engineering platform yang tidak
mengizinkan akses (bandingkan dengan larangan YouTube/Spotify di
CLAUDE.md: itu soal *mengambil konten yang tidak dibeli/dilisensikan
user*, sedangkan ini murni akses ke pembelian sah milik user sendiri
lewat API yang disediakan Bandcamp untuk tujuan itu).

Pola credential-nya sama seperti refresh token Drive di ADR 0002: didapat
manual di luar aplikasi (login ke bandcamp.com, salin cookie `identity`
dari browser), ditempel ke form admin, dienkripsi sebelum disimpan.
Kredensial disimpan sebagai JSON `{identity_cookie, fan_id}` di dalam
`credentials_encrypted`.

**Batasan scope v1 yang disengaja**: item pembelian tipe "track" (single
file) didukung penuh. Item tipe "album" biasanya mengunduh sebagai .zip
multi-track dari Bandcamp — ekstraksi zip + split metadata per-track
TIDAK dibangun di Sprint 10 (nambah kompleksitas signifikan: penentuan
track number dari nama file di dalam zip, retry per-track, dll). Item
album di-skip dengan log jelas ("album downloads not supported yet"),
bukan bikin gagal seluruh sync run. Kalau ini penting nanti, jadikan
sprint/task terpisah alih-alih diselipkan di sini.

### 5. Cloud sync: HANYA Dropbox, bukan tiga provider sekaligus

CLAUDE.md menyebut "Dropbox/OneDrive/iCloud" sebagai contoh, bukan
mandat membangun ketiganya. Dropbox dipilih sebagai satu-satunya
implementasi konkret untuk `cloud_sync` karena API-nya bersih, publik,
dan pakai OAuth refresh token (pola identik dengan Drive). OneDrive
(butuh Azure AD app registration) dan iCloud (tidak punya API publik
sama sekali) masuk backlog — bukan penyimpangan spec, karena
`ingest_jobs.source_type` tetap generik `'cloud_sync'` (implementasi di
baliknya bisa diperluas nanti tanpa migrasi skema baru).

### 6. Sync idempoten lewat checksum dedup yang sudah ada, bukan tabel "seen items" baru

`Sync` memanggil client (`ListNewPurchases`/`ListNewFiles`) dengan
parameter `since = last_synced_at`, lalu untuk tiap item: download ke
`INGEST_TMP_DIR`, hitung checksum, panggil jalur ingest yang sama persis
dengan upload manual (`ingest.Service.Accept` yang sekarang menerima
`sourceType`, lalu `Process` yang sudah ada dari Sprint 3). Kalau
checksum sudah ada, `Accept` otomatis short-circuit ke `completed` tanpa
proses ulang (perilaku ini sudah ada sejak Sprint 3) — jadi retry/re-sync
yang tumpang tindih aman tanpa perlu tabel pelacakan "item mana yang
sudah diproses" yang terpisah.

## Konsekuensi

- `ingest.Service.Accept` sekarang menerima parameter `sourceType`
  (sebelumnya hardcoded `"manual_upload"`) — satu-satunya call site HTTP
  upload diupdate untuk mengirim `"manual_upload"` eksplisit, perilaku
  tidak berubah untuk jalur itu.
- **Butuh input manual dari user** (sama seperti Drive di Sprint 3/4):
  cookie `identity` + `fan_id` Bandcamp asli, dan `refresh_token` OAuth
  Dropbox asli, keduanya belum ada. Kode connect/sync lengkap dan gagal
  dengan pesan jelas tanpa kredensial asli (pola yang sama persis dengan
  Drive), bukan crash.
- Kalau nanti benar-benar perlu OneDrive/iCloud, tambahkan implementasi
  client baru di belakang `source_type = 'cloud_sync'` yang sama — tidak
  perlu migrasi skema baru.
