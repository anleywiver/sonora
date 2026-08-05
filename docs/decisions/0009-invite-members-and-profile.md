# ADR 0009 — Invite Members, Profile, Browse Library (Sprint 14 sisipan)

Status: Diterima
Tanggal: Sprint 14 (sisipan kedua, atas permintaan user)

## Konteks

User meminta 5 hal sekaligus: WhatsApp "request access" link di Login,
halaman Profile, halaman Browse Library (mengganti alias Library →
Favorite), halaman Admin Login terpisah, dan halaman Admin Settings.
Permintaan juga menyebut `docs/screens-spec.md` sudah di-update dengan
spec detail — nyatanya belum (dicek langsung, file tetap 60 baris seperti
sebelumnya). Bagian ini didesain dari deskripsi prosa user secara
langsung, ditambahkan ke screens-spec.md sebagai bagian dari pekerjaan ini
(bukan diasumsikan sudah ada).

Bagian "Invite Member" (dari permintaan sebelumnya, admin/users) butuh
keputusan arsitektur baru yang tidak trivial, dijelaskan di sini.

## Keputusan

### Invite = pre-create user row dengan `google_id` NULL, di-klaim saat login pertama

Sistem ini TIDAK punya infrastruktur email (tidak ada SMTP/email service
di project ini sama sekali) — "invite" tidak mengirim email beneran.
Sebagai gantinya: `POST /admin/users/invite` membuat baris `users` dengan
`email` terisi, `google_id = NULL` ("pending"), `role = 'member'`. Saat
orang itu benar-benar login lewat Google OAuth, `HandleGoogleCallback`
sekarang cek dulu apakah ada baris pending dengan email yang cocok
SEBELUM membuat user baru — kalau ada, baris itu di-"claim" (isi
`google_id`/`name`/`avatar_url` dari profil Google), bukan bikin user
kedua. Ini konsisten dengan model akses "invite-only lewat Owner secara
personal" (bukan self-signup) yang sudah jadi asumsi implisit sejak
Sprint 2.

`google_id` diubah jadi nullable (migration 000009) — NULL, bukan string
kosong `""`, supaya unique index tetap benar (Postgres tidak menganggap
banyak NULL sebagai duplikat, beda dengan string kosong yang akan
bentrok di invite kedua).

### Status user di admin UI: "Active" vs "Invited"

Dihitung dari `google_id IS NULL` (belum pernah login) vs tidak — bukan
kolom status terpisah, supaya tidak ada dua sumber kebenaran untuk hal
yang sama.

### Hapus akses Member TIDAK PERNAH bisa untuk Owner

`DELETE /admin/users/:id` menolak (403) kalau target adalah Owner —
mencegah admin yang sedang login menghapus dirinya sendiri atau (di
skenario multi-admin masa depan) Owner lain, konsisten dengan asumsi
single-owner personal deployment.

## Konsekuensi lain (fitur non-users di sisipan yang sama)

- **WhatsApp request-access link**: murni link `wa.me` client-side, TIDAK
  ada endpoint/tabel baru — sesuai instruksi eksplisit user (tetap
  invite-only, ini cuma shortcut kontak).
- **Profile**: `PUT /auth/me` baru (update `name`/`avatar_url`), avatar
  upload pakai jalur storage yang sama dengan ingest (checksum + upload
  ke storage account aktif) — bukan jalur baru.
- **Browse Library**: 3 endpoint baru (`GET /library/songs`,
  `/library/albums`, `/library/artists`) — beda dari `/favorites` (yang
  cuma nunjukin item yang di-like): ini nunjukin SEMUA koleksi (semua
  lagu yang pernah di-ingest, bukan cuma favorit).
- **Admin Login/Access Denied**: halaman terpisah eksplisit (bukan cuma
  gate diam-diam di `app-shell.tsx` yang sudah ada sejak Sprint 9) —
  pesan jelas ke Member yang salah masuk, bukan redirect senyap.
- **Admin Settings**: `app_settings` singleton table baru (app name,
  default language, maintenance mode) — bukan per-user, satu baris
  global untuk seluruh deployment personal ini.
