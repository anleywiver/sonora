# ADR 0012 — Credential-Based Auth, Google OAuth as a Runtime Toggle

Status: Diterima
Tanggal: Sprint 14 (sisipan keempat, atas permintaan user)

## Konteks

User meminta pivot besar: username/password jadi metode login DEFAULT,
Google OAuth (Sprint 2, satu-satunya metode sejak awal project) TETAP
ADA sepenuhnya di kode dan berfungsi, tapi bisa dinyalakan/dimatikan
lewat toggle di admin panel saat runtime — bukan dihapus, bukan
di-hardcode nonaktif secara permanen. Migration diberi nomor `000011`
(bukan `000008` seperti disebut user — `000008` sudah dipakai Sprint 14
sisipan pertama untuk `ingest_filter_rules`, sudah di-commit).

## Keputusan

### Toggle disimpan di `app_settings`, dicek di backend (bukan cuma disembunyikan di frontend)

`app_settings` (key-value store baru) menyimpan `google_oauth_enabled`.
`GET/POST /auth/google*` TETAP terdaftar dan berfungsi penuh di
`main.go` — tidak dihapus, tidak dikondisikan lewat build flag. Yang
baru: keduanya cek `app_settings.google_oauth_enabled` di AWAL handler,
balas 403 kalau `false`. Ini defense-in-depth eksplisit yang diminta
user: frontend menyembunyikan tombol Google berdasarkan `/auth/config`,
TAPI backend menolak juga kalau seseorang memanggil endpoint itu
langsung meski tombolnya disembunyikan.

### Dua endpoint login kredensial, bukan satu

`POST /auth/login/admin` (email+password, HARUS role owner) dan
`POST /auth/login` (username+password, role member) dipisah secara
eksplisit — user Owner tidak bisa login lewat endpoint member dan
sebaliknya, meniru pemisahan yang sudah ada di admin app (login
terpisah dari user app) tapi di level backend juga, bukan cuma UI.
Keduanya rate-limited 5 req/menit per IP (lebih ketat dari limiter
global 300/menit Sprint 12 — endpoint credential-guessing paling
realistis untuk brute-force).

### Dua jalur pembuatan Member: invite-by-email (Google) TETAP ADA, ditambah create-by-password (baru)

`POST /admin/users/invite` (ADR 0009, Sprint 14 sisipan sebelumnya)
TIDAK dihapus — itu masih jalur yang benar untuk Member yang login
lewat Google. Ditambahkan jalur BARU: form "Add User" (username, nama,
password manual/generate) yang langsung aktif tanpa perlu klaim — cocok
untuk Member yang login lewat kredensial. Owner memilih jalur mana yang
relevan tergantung `google_oauth_enabled` saat itu.

### Bootstrap Owner pertama: CLI terpisah, bukan lewat endpoint

`cmd/seed-owner` — dijalankan manual (`go run ./cmd/seed-owner --email=...
--password=...`), bcrypt cost 12. TIDAK ada endpoint HTTP untuk membuat
Owner pertama — kalau ada, siapa pun yang mengakses API sebelum ada Owner
bisa membuat dirinya sendiri jadi Owner. Pola ini sama dengan alasan
"user pertama otomatis jadi Owner" di Sprint 2 (find-or-create via Google)
kini digantikan/didampingi jalur CLI eksplisit untuk kredensial.

## Konsekuensi

- `users.email` tetap `NOT NULL UNIQUE` (skema asli Sprint 2, sengaja
  TIDAK diubah — resiko terlalu besar untuk fitur sisipan). Form "Add
  User" (username/password) tidak minta email dari Owner; kalau
  dikosongkan, backend generate placeholder `username@local.invalid`
  supaya constraint tetap terpenuhi tanpa memaksa Owner mengarang email
  asli untuk Member yang memang tidak akan pernah login lewat Google.
- `users.username`/`password_hash` nullable — user Google-only tidak
  pernah punya password; user kredensial-only tidak perlu klaim Google.
- Karena `app_settings.google_oauth_enabled` di-seed `false` (sesuai
  instruksi), TIDAK ADA cara login sampai `cmd/seed-owner` dijalankan
  manual pertama kali — ini disengaja (mencegah akses tanpa credential
  sama sekali sebelum Owner benar-benar di-bootstrap).
- `GET /auth/config` sengaja PUBLIC (tanpa auth) — halaman Login perlu
  tahu tombol apa yang ditampilkan SEBELUM ada token sama sekali.
