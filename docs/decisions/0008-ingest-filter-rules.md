# ADR 0008 — Ingest Filter Rules for Auto-Ingest Sources

Status: Diterima
Tanggal: Sprint 14 (sisipan, atas permintaan user)

## Konteks

User meminta filter genre/tahun untuk auto-ingest dari Bandcamp/cloud
sync, supaya item yang tidak sesuai kriteria tidak menghabiskan resource
(upload storage, waveform generation, MusicBrainz lookup) atau kuota
storage secara sia-sia. Beberapa detail teknis dari permintaan awal perlu
disesuaikan terhadap kode yang sudah ada:

1. Migration diberi nomor `000008` (bukan `000007` seperti diminta) —
   `000007` sudah dipakai Sprint 14 untuk `lyrics_provider_stats`.
2. `ingest_jobs_status_check` TETAP pakai `'processing'` (bukan diganti
   `'running'` seperti SQL awal) — mengganti nama itu akan merusak
   `MarkIngestJobProcessing` yang sudah dipakai semua job sejak Sprint 3.
   Dua status baru (`needs_manual_upload`, `skipped_by_filter`)
   ditambahkan ke enum yang ada, bukan menggantikannya.
3. Tidak ada tabel `ingest_job_logs` terpisah — alasan skip disimpan di
   `ingest_jobs.error_message` yang sudah ada (kolom yang sama dipakai
   untuk pesan error job `failed`), bukan tabel baru untuk satu string.

## Keputusan

### Genre/tahun didapat dari ffprobe, BUKAN dari listing Bandcamp/Dropbox

`bandcamp.Purchase` dan `dropbox.File` (Sprint 10) tidak membawa genre/
tahun — data itu cuma ada di tag ID3 file audio-nya sendiri. Jadi filter
TIDAK bisa dicek sebelum download (di titik "item terdeteksi di
Bandcamp/cloud sync" seperti diminta) — cek baru bisa terjadi setelah
file di-download ke temp dan `mediainfo.Probe` (ffprobe) jalan, yang
sudah terjadi di awal `ingest.Process`. `mediainfo.Info` diperluas dengan
field `Genre` dan `Year` (parse tag `genre` dan `date`/`year` ID3 yang
ffprobe sudah expose, tidak perlu tool baru).

Filter check ditempatkan tepat setelah `Probe` berhasil, SEBELUM
`uploadToStorage`/waveform/MusicBrainz — jadi tujuan "hemat resource &
storage quota" tetap tercapai meski titik cek-nya bukan sebelum-download
(karena secara teknis tidak mungkin sebelum-download dengan data yang
tersedia).

### Genre/tahun yang hilang TIDAK memblokir (fail-open, bukan fail-closed)

Kalau file tidak punya tag genre/tahun sama sekali tapi ada rule yang
mensyaratkannya, item itu LOLOS (bukan otomatis di-skip). Alasan: metadata
ID3 yang hilang/kosong adalah hal biasa (banyak rip legal tanpa tag
lengkap), dan memblokir berdasarkan ketiadaan data berisiko lebih besar
(kehilangan lagu yang sah) dibanding risiko sebaliknya (satu lagu di luar
kriteria lolos). Filter cuma aktif kalau datanya ADA dan tidak cocok.

### Semantik rule

- `genre_allow` — allow-list: kalau ADA minimal satu rule `genre_allow`
  untuk source_type itu, genre lagu (case-insensitive) harus cocok salah
  satu, atau di-skip. Kalau TIDAK ADA rule `genre_allow` sama sekali,
  semua genre lolos (default permissive, bukan default deny).
- `year_min`/`year_max` — lagu harus rilis di antara batas itu (inklusif)
  kalau rule-nya ada. Keduanya independen (boleh cuma salah satu).

### Scope: `needs_manual_upload` masih reserved, belum ada logic

Ditambahkan ke CHECK constraint sesuai permintaan, tapi belum ada kode
yang benar-benar men-set status ini — tidak jelas dari permintaan awal
kapan itu dipicu. Dicatat sebagai reserved value untuk sekarang, bukan
fitur yang setengah jalan disembunyikan.

## Konsekuensi

- `application/ingestfilter` — service CRUD rules + `Check(sourceType,
  genre, year) (pass bool, reason string)`.
- `ingest.Process` sekarang butuh tahu `SourceType` job (sudah ada di
  row) untuk decide apakah perlu filter check sama sekali — manual_upload
  selalu skip pengecekan ini tanpa exception.
- Admin endpoint baru: `GET/POST/DELETE /admin/ingest-sources/:source_type/filters[/:id]`.
