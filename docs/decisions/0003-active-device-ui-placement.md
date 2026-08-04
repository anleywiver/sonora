# ADR 0003 — Active Device UI Placement & Device Identity Plumbing

Status: Diterima
Tanggal: Sprint 8

## Konteks

Sprint 8 ("Active Device + PWA") mengimplementasikan Transfer Playback —
tapi `docs/screens-spec.md` tidak pernah mendesain layar atau komponen
untuk memilih/transfer Active Device sama sekali. Bottom utility row Now
Playing di screens-spec cuma menyebut "Sleep Timer, Speed, Lyrics, Queue",
tidak ada "Devices"/"Connect".

Selain itu, ditemukan gap plumbing: JWT access token hanya bawa
`user_id`+`role` (sengaja, per konvensi Sprint 2 — access token tidak
terikat device). Tapi frontend butuh tahu "device_id saya sendiri" untuk
Active Device (bandingkan dengan `playback_state.active_device_id`, kirim
ke `/ws/token`). Sebelum Sprint 8, `device_id` yang dibuat saat OAuth
callback / refresh tidak pernah dikembalikan ke frontend sama sekali.

## Keputusan

1. **Device identity ke frontend**: `TokenPair` (application/auth)
   ditambah field `DeviceID`. `POST /auth/refresh` dan redirect
   `/auth/callback` (fragment URL) sekarang menyertakan `device_id`.
   Frontend menyimpannya di `store/auth.ts` bareng access token (di
   memori, sama seperti access token — bukan sesuatu yang sensitif untuk
   disimpan tapi konsisten dengan access token yang juga direset tiap
   full reload, dan device_id akan didapat lagi dari `/auth/refresh`).

2. **UI Transfer Playback**: ditaruh sebagai tombol "Devices" baru di
   bottom utility row Now Playing (sejajar Lyrics/Queue) yang membuka
   bottom sheet list device (`GET /devices` + highlight yang aktif),
   tap salah satu → `POST /player/transfer`. Pola ini meniru "Connect to
   a device" Spotify Connect yang sudah jadi referensi desain sejak awal
   (`CLAUDE.md` "Active Device pattern... mirip Spotify Connect").

## Konsekuensi

- `docs/screens-spec.md` TIDAK diubah (dokumen desain asli dari sesi
  perencanaan tetap sebagai referensi historis) — halaman/komponen baru
  ini dicatat di sini dan di `CLAUDE.md`, bukan menyisipkan ke spec lama.
- Sprint 14 (UI Fidelity) perlu memperhitungkan komponen ini saat review
  akhir karena tidak ada di checklist 21 halaman asli.
