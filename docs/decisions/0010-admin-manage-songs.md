# ADR 0010 — Admin Manage Songs (Sprint 14 sisipan)

Status: Diterima
Tanggal: Sprint 14 (sisipan)

## Konteks

Halaman admin baru untuk mengelola katalog lagu: list+search, edit
metadata (title/artist/album/genre), dan hapus (soft delete). "Artist"
dan "album" bukan kolom teks bebas di `songs` — keduanya foreign key ke
`artists`/`albums`. "Genre" bukan kolom sama sekali — relasi many-to-many
lewat `song_genres`. Edit metadata butuh find-or-create yang sama seperti
pipeline ingest (Sprint 3/4), bukan update kolom langsung.

## Keputusan

### Edit metadata pakai find-or-create yang sama seperti ingest

`PATCH /admin/songs/:id` menerima `title`/`artist_name`/`album_title`/
`genre_name` opsional. `artist_name`/`album_title` di-resolve lewat
find-or-create (persis pola `findOrCreateArtist`/`findOrCreateAlbum` di
`application/ingest`) — kalau nama artist/album itu sudah ada, dipakai
lagi; kalau belum, dibuat baru. `genre_name` mengganti SELURUH relasi
`song_genres` lagu itu jadi satu genre itu saja (UI menyebut "genre"
tunggal, bukan multi-select) — bukan menambah ke daftar genre yang ada.

### Soft delete: scope yang benar-benar dikerjakan vs yang didokumentasikan sebagai belum

`songs.deleted_at` (migration `000010`) ditambahkan. Yang BENERAN
dikerjakan sprint ini:
- Lagu dengan `deleted_at` terisi hilang dari list admin Songs.
- Lagu itu di-`DeleteDocuments` dari index Meilisearch (jadi tidak
  muncul lagi di hasil `/search` user).

Yang SENGAJA belum dikerjakan (dicatat di sini, bukan disembunyikan):
`GetSong`/`GetSongDetail`/stream endpoint TIDAK dicek terhadap
`deleted_at` — kalau lagu itu masih ada di playlist/favorite/queue user
lain, mereka masih bisa memutarnya lewat link yang sudah ada. Menutup
celah itu penuh (cek `deleted_at` di setiap query catalog, playlist,
favorite, history, queue) adalah pekerjaan jauh lebih besar dari yang
diminta ("table + tombol hapus dengan konfirmasi") — kalau nanti
memang perlu enforcement penuh, itu layak jadi task terpisah, bukan
disisipkan diam-diam di sini dengan cakupan yang mengecoh.

## Konsekuensi

- `application/adminsongs` — service baru, terpisah dari
  `application/catalog` (yang tetap murni read-path untuk user biasa)
  supaya logic admin-only (soft delete, override metadata) tidak
  bercampur dengan jalur baca yang sudah stabil sejak Sprint 4.
- Endpoint baru: `GET /admin/songs` (search+cursor pagination),
  `PATCH /admin/songs/:id`, `DELETE /admin/songs/:id`.
