# ADR 0005 — Waveform, MusicBrainz Enrichment, Admin Analytics (Sprint 11)

Status: Diterima
Tanggal: Sprint 11

## Konteks

Roadmap Sprint 11: "Polish ingest — waveform generation, metadata
MusicBrainz, admin Analytics." Seperti Sprint 10, `docs/api-design.md`
tidak punya spesifikasi skema/response untuk waveform atau MusicBrainz —
kecuali dua endpoint analytics yang sudah ada di baris Admin
(`/admin/analytics/top-played`, `/admin/analytics/storage-growth`).

## Keputusan

### 1. Waveform: peak array dari ffmpeg, bukan library baru

Worker sudah punya `ffmpeg` (Sprint 3, untuk `ffprobe`). Waveform
dihasilkan dengan decode ke PCM mono 8-bit mentah lewat
`ffmpeg -i <path> -ac 1 -filter:a aresample=8000 -f u8 pipe:1`, lalu Go
membagi byte stream itu jadi ~200 bucket dan ambil peak (`|sample-128|`)
per bucket — hasil array `[]int16` (0-255) disimpan di kolom baru
`songs.waveform_peaks SMALLINT[]`. Tidak menambah dependency baru (tidak
pakai library waveform pihak ketiga), konsisten dengan pola `mediainfo`
yang sudah ada.

Generate waveform berjalan di `ingest.Process` setelah `ffprobe`, SEBELUM
job ditandai `completed`, tapi kegagalannya **tidak fatal** — log lalu
lanjut dengan `waveform_peaks = NULL` (sama seperti index Meilisearch di
Sprint 4: song row tetap sumber kebenaran yang utuh).

### 2. MusicBrainz: enrichment best-effort, rate-limited, TIDAK fatal

MusicBrainz API publik (`musicbrainz.org/ws/2`) tidak butuh API key, tapi
kebijakan penggunaannya mensyaratkan rate limit 1 request/detik dan
`User-Agent` deskriptif — keduanya diterapkan di
`infrastructure/musicbrainz` (rate limiter package-level pakai
`golang.org/x/time/rate`, sudah jadi transitive dependency lewat
`google.golang.org/api`, jadi tidak nambah entry `go.mod` baru).

Alur: setelah `CreateSong`, cari recording berdasarkan title+artist+
duration (toleransi ±2 detik). Kalau ketemu, simpan `musicbrainz_id` di
`songs`/`artists`/`albums` (kolom baru, nullable) dan — HANYA kalau
`cover_url`/`image_url` masih kosong — ambil cover art dari Cover Art
Archive (`coverartarchive.org`, juga publik tanpa key) memakai release
MBID. Gagal di titik manapun (tidak ketemu, API error, rate limit) diam-
diam di-log dan diabaikan — song tetap `completed` dengan metadata
ffprobe apa adanya, TIDAK memblokir ingest.

**Beda dengan Bandcamp/Dropbox di Sprint 10**: MusicBrainz + Cover Art
Archive tidak butuh credential apapun — jadi ini satu-satunya bagian
Sprint 11 yang bisa diverifikasi end-to-end penuh tanpa "butuh input
manual dari user".

### 3. Admin Analytics: dua query baru, chart pakai `recharts` yang sudah ada

`recharts` sudah jadi dependency `apps/admin` sejak Sprint 1 scaffold
(belum pernah dipakai). `GET /admin/analytics/top-played` (top 10 lagu by
play count dari `play_history`) dan `GET /admin/analytics/storage-growth`
(total bytes storage per bulan, 6 bulan terakhir, dari `storage_files.created_at`)
diimplementasi sebagai sqlc query baru + handler admin, di-render di
`/analytics` (sebelumnya placeholder sejak Sprint 9) pakai bar chart
`recharts` sesuai `docs/screens-spec.md` #21.

## Konsekuensi

- Migration `000006`: `songs.waveform_peaks SMALLINT[]`,
  `songs.musicbrainz_id`, `artists.musicbrainz_id`,
  `albums.musicbrainz_id` (semua nullable TEXT/array, tidak mengubah
  constraint yang ada).
- `ingest.Process` sedikit lebih lama per job (waveform decode + 1-2
  network call MusicBrainz/Cover Art dengan rate limit 1/detik) — dampak
  minor untuk skala personal, dicatat kalau nanti jadi masalah nyata di
  Sprint 13 (load test ringan).
- TIDAK ada fallback lain kalau MusicBrainz tidak menemukan match (mis.
  Discogs, Last.fm) — satu provider metadata cukup untuk skala personal,
  sesuai filosofi "jangan over-engineer".
