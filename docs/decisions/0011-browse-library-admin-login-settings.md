# ADR 0011 — Browse Library, Admin Login, Admin Settings (Sprint 14 sisipan)

Status: Diterima
Tanggal: Sprint 14 (sisipan ketiga)

## Konteks

Tiga fitur tersisa dari permintaan sisipan: halaman Browse Library (semua
koleksi, bukan cuma favorit), halaman Admin Login terpisah dengan
Access Denied eksplisit, dan halaman Admin Settings (app name, bahasa,
maintenance mode).

## Keputusan

### Browse Library: TIDAK pakai cursor pagination

`GET /library/songs|albums|artists` sengaja TIDAK pakai keyset/cursor
pagination seperti kebanyakan endpoint list lain di API ini — cukup
`LIMIT` datar (200) + search + toggle sort (`recent` vs `alpha`).
Alasannya murni skala: koleksi personal realistisnya ratusan item, bukan
jutaan — butuh cursor pagination PER MODE sort (recent butuh cursor
`created_at`, alpha butuh cursor `title`) adalah kompleksitas nyata untuk
manfaat yang tidak terasa di skala ini. Beda dengan `/admin/jobs`
misalnya, yang datanya memang bisa tumbuh tanpa batas seiring waktu.

### Maintenance mode: dicek di tiap request lewat middleware baru, bukan di setiap handler

`app_settings.maintenance_mode` (tabel singleton baru, 1 baris) dicek
lewat middleware baru `middleware.MaintenanceGate` yang jalan SETELAH
`RequireAuth` (butuh tahu role) tapi SEBELUM handler — Member yang kena
maintenance dapat `503` dengan pesan jelas, Owner tetap lolos. Middleware
dipasang di level route group `api` (semua endpoint user-facing), BUKAN
di `adminGroup` (Owner harus tetap bisa matikan maintenance mode dari
admin meski mode itu aktif).

### Admin Login: halaman terpisah, Access Denied eksplisit bukan redirect

`apps/admin/src/app/login/page.tsx` sudah ada sejak Sprint 9 (Google-only,
tanpa email/password) — tidak perlu dibangun ulang, cuma dikonfirmasi
sudah sesuai spec sisipan ini (tidak ada request-access link, itu cuma
untuk app user biasa). Yang baru: `app-shell.tsx` sejak Sprint 9 sudah
py gate "Access Denied" untuk role bukan Owner — dikonfirmasi ulang
sudah sesuai maksud sisipan ini (pesan jelas, bukan redirect diam-diam),
bukan dibangun dari nol.

## Konsekuensi

- Migration baru: `app_settings` (singleton, 1 baris di-seed lewat
  migration itu sendiri supaya GET pertama tidak pernah 404).
- `GET/PATCH /admin/settings` — Owner only, sama seperti admin lain.
- Browse Library TIDAK exhaustive untuk koleksi sangat besar (ribuan+
  lagu) — kalau nanti jadi masalah nyata, itu layak jadi task terpisah
  (tambah cursor pagination per sort mode), bukan dipaksakan sekarang.
