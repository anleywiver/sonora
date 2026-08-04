# Screen Specification — Sonora

21 halaman total (15 user + 6 admin section). Checklist elemen wajib per
halaman — pakai `docs/design-system.md` untuk styling detail tiap elemen.
Semua halaman **mobile-first**, dark mode only.

## User-facing (apps/frontend)

**1. Splash Screen** — Logo (headphone icon di gradient box), blue radial glow di belakang (animasi pulse pelan), loading dots (3 dot bounce sequential), auto-navigate ke Login/Home setelah ±1.5 detik.

**2. Login** — Logo kecil + tagline, tombol Google (putih solid, branding asli Google — JANGAN ubah warna), tombol Apple (placeholder, styling redup/disabled), divider "or continue with email", input email+password, tombol Sign in (primary).

**3. Home** — Header (greeting + avatar), Continue Listening (progress bar mini), Recently Played (horizontal scroll), Trending (card dengan artwork+judul+artist), New Release (badge "NEW"), Favorite Artists (avatar bulat horizontal), Quick Mix (gradient card 2 kolom), Bottom Navigation + Mini Player menempel.

**4. Search** — Search bar, Recent Searches (dengan tombol Clear), Trending Searches (pill/chip), Browse Genre (grid 2 kolom gradient card).

**5. Search Result** — Filter tab (All/Songs/Artists/Lyrics), Top Result (card besar), Songs list (artwork+title+artist+duration+heart), Lyrics Match (cuplikan teks + info lagu).

**6. Song Detail** — Hero gradient (biru dari artwork), artwork besar center, judul+artist+album, quick action (shuffle, play besar, download), bottom action row (Like, Queue, Lyrics, Share).

**7. Now Playing** — Radial glow background, artwork besar, title+artist+like, seekbar+waktu, kontrol utama (shuffle, prev, play/pause besar putih, next, repeat), bottom utility row (Sleep Timer, Speed, Lyrics, Queue).

**8. Lyrics (fullscreen)** — Header (chevron down, mini info lagu, dots menu), lirik center dengan highlight baris aktif (bg accent/15) + fade progresif baris lain, footer (tap-to-seek hint, auto-scroll toggle).

**9. Queue** — Now Playing highlight (bg accent/10), "Next up" list dengan drag handle (grip-vertical icon) + tombol remove (X), "Clear queue" di header.

**10. Playlist Detail** — Hero gradient dari cover, cover besar, nama+deskripsi+jumlah lagu, action row (play besar, shuffle, like, search-in-playlist, sort), tracklist bernomor.

**11. Favorite** — Tab kategori (Songs/Albums/Artists/Playlists), list sesuai tab aktif dengan heart-filled icon.

**12. Artist Detail** — Banner gradient + avatar overlap (border background color), nama+monthly listeners, tombol Play+Follow, Popular songs list, Related artists (avatar horizontal).

**13. Album Detail** — Hero gradient dari cover, cover+judul+artist+tahun+genre+jumlah track, action row (play, shuffle, download), tracklist bernomor dengan heart per track.

**14. Downloads** — Filter tab (All/Downloading/Failed), In Progress (progress bar + persen), Failed (merah + tombol retry), Available Offline (centang hijau + ukuran file).

**15. Settings** — Grouped list: Appearance (Theme), Storage (usage bar + Connected Google Drive dengan badge jumlah akun + Clear cache), Playback (Lyrics source, Language), About (versi app).

## Admin (apps/admin)

**16. Dashboard** — Sidebar nav (Dashboard, Drive Manager, Ingest Sources*, Lyrics Source, Job Queue, Analytics), stat card 4 kolom (Total song/user/storage/drive), Storage Distribution (bar per drive dengan warna warning kalau >90%), Background Jobs summary, Top Played list.

> *Catatan penting*: section ini di desain awal namanya "Crawler" — **ganti nama jadi "Ingest Sources"** dan isinya list provider legal (Bandcamp, Cloud Sync), BUKAN keyword/auto-download generic crawler. Ini koreksi dari STEP 2 keputusan (Auto Ingest legal-only).

**17. Drive Manager** — Card per storage account (nama, badge health status, progress bar quota, sisa GB), tombol "+ Add drive" dan "Run health check".

**18. Ingest Sources** (dulu "Crawler") — List provider aktif (Manual Upload, Bandcamp, Cloud Sync) dengan status connected/disconnected, BUKAN form keyword/auto-download.

**19. Lyrics Source** — Table provider dengan drag-handle untuk reorder priority, kolom Health (Online/Rate limited), kolom Match Rate.

**20. Job Queue** — Table dense (Job, Type, Status badge, Action) untuk `ingest_jobs`, tombol Retry untuk yang failed.

**21. Analytics** — Storage Growth (bar chart 6 bulan), Download Trend + Most Played (2 kolom kecil berdampingan).

## Interaksi Wajib (semua halaman)

- Page transition: fade + slight scale (Framer Motion)
- Hero animation: shared layout transition artwork dari list → detail (pakai `layoutId` Framer Motion)
- Pull-to-refresh di Home (opsional, nice-to-have)
- Skeleton loading state untuk semua list yang fetch data (jangan blank/spinner polos)
