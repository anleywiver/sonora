# Screen Specification — Sonora

21 halaman inti (15 user + 6 admin section) dari desain awal, plus 7
halaman/section sisipan Sprint 14 (#22–28, lihat bagian "Tambahan Screen"
di bawah). Checklist elemen wajib per halaman — pakai
`docs/design-system.md` untuk styling detail tiap elemen. Semua halaman
**mobile-first**, dark mode only.

## User-facing (apps/frontend)

**1. Splash Screen** — Logo (headphone icon di gradient box), blue radial glow di belakang (animasi pulse pelan), loading dots (3 dot bounce sequential), auto-navigate ke Login/Home setelah ±1.5 detik.

**2. Login** — Logo kecil + tagline. Form username+password (primary, SELALU ditampilkan — Sprint 14 sisipan, ADR 0012) + tombol Sign in. Tombol Google (putih solid, branding asli Google) dan tombol Apple (placeholder, redup/disabled) HANYA muncul kalau `GET /auth/config` bilang `google_oauth_enabled=true` — dicek saat halaman dimuat, bukan hardcoded tampil. Link teks kecil "Belum punya akses? Request akses" ke WhatsApp Owner (`NEXT_PUBLIC_OWNER_WHATSAPP`, sisipan Batch 2) — shortcut kontak, BUKAN form registrasi/self-signup.

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

**15. Settings** — Grouped list: Appearance (Theme), Storage (usage bar + Connected Google Drive dengan badge jumlah akun + Clear cache), Playback (Lyrics source, Language), About (versi app), link ke Profile.

## Admin (apps/admin)

**16. Dashboard** — Sidebar nav (Dashboard, Drive Manager, Ingest Sources*, Lyrics Source, Job Queue, Analytics), stat card 4 kolom (Total song/user/storage/drive), Storage Distribution (bar per drive dengan warna warning kalau >90%), Background Jobs summary, Top Played list.

> *Catatan penting*: section ini di desain awal namanya "Crawler" — **ganti nama jadi "Ingest Sources"** dan isinya list provider legal (Bandcamp, Cloud Sync), BUKAN keyword/auto-download generic crawler. Ini koreksi dari STEP 2 keputusan (Auto Ingest legal-only).

**17. Drive Manager** — Card per storage account (nama, badge health status, progress bar quota, sisa GB), tombol "+ Add drive" dan "Run health check".

**18. Ingest Sources** (dulu "Crawler") — List provider aktif (Manual Upload, Bandcamp, Cloud Sync) dengan status connected/disconnected, BUKAN form keyword/auto-download.

> *Catatan Sprint 14 sisipan (ADR 0008)*: tambah panel "Filter Rules" per source (Bandcamp, Cloud Sync) — genre allow-list (chip, tambah/hapus) dan year range (min/max). Filter ini **HANYA berlaku untuk auto-ingest** (Bandcamp/cloud sync), **TIDAK PERNAH** untuk Manual Upload — dijelaskan eksplisit di UI, bukan cuma diasumsikan.

**19. Lyrics Source** — Table provider dengan drag-handle untuk reorder priority, kolom Health (Online/Rate limited), kolom Match Rate.

**20. Job Queue** — Table dense (Job, Type, Status badge, Action) untuk `ingest_jobs`, tombol Retry untuk yang failed.

**21. Analytics** — Storage Growth (bar chart 6 bulan), Download Trend + Most Played (2 kolom kecil berdampingan).

## Tambahan Screen (sisipan Sprint 14, atas permintaan user)

**22. Users** — Table (Nama, Email, Role, Status, Join date), tombol "Invite Member" (modal input email), tombol hapus akses per Member (tidak ada untuk Owner). "Invite" TIDAK kirim email beneran (tidak ada infrastruktur email di project ini) — baris user dibuat dengan status "Invited", otomatis jadi "Active" begitu email itu login lewat Google pertama kali (lihat ADR 0009).

**23. Songs** — Table semua lagu (judul, artist, album, durasi, storage provider, tanggal ditambahkan), search/filter by judul/artist, tombol edit metadata (title/artist/album/genre), tombol hapus (soft delete + konfirmasi).

## Tambahan Screen (sisipan kedua Sprint 14, apps/frontend)

**24. Profile** — Avatar (tap untuk ganti — resize client-side ke thumbnail 128px lalu kirim sebagai `data:image/...`, BUKAN upload ke storage pool Drive, lihat ADR 0009), form nama (editable, `PUT /auth/me`), email (read-only), badge role, tombol Sign out.

**25. Browse Library** — Tab (Songs/Albums/Artists/Playlists), search + sort (A–Z/Terbaru) per tab. Beda dari Favorite (#11): menampilkan SELURUH katalog, bukan cuma yang di-favorite (ADR 0011 — flat `LIMIT 200`, bukan cursor pagination, cukup untuk skala koleksi personal).

## Tambahan Screen (sisipan ketiga Sprint 14, apps/admin — ADR 0012)

**26. Admin Login** — Sama pola dengan Login user-facing (#2): form email+password (primary, selalu tampil), tombol Google muncul kalau `google_oauth_enabled=true`. Sudah ada strukturnya sejak Sprint 9 (`app-shell.tsx` gate) — sisipan ini menambahkan form kredensial di atasnya.

**27. Admin Access Denied** — Bukan halaman terpisah, tapi state dari `AppShell`: kalau role user yang login bukan `owner`, sidebar+children diganti pesan "Access Denied" eksplisit (bukan redirect diam-diam atau 403 mentah) — sudah ada sejak Sprint 9, diverifikasi ulang di sisipan ini.

**28. Admin Settings** — App Name (input+save), Default Language (pill Indonesia/English), toggle "Enable Google OAuth Login" (langsung PATCH `/admin/settings`, real-time, tanpa restart — mengontrol tombol Google di #2 dan #26 SEKALIGUS backend `/auth/google*`), toggle "Maintenance Mode" (503 untuk non-Owner di semua endpoint API non-auth/non-admin). Link "Buka Drive Manager" untuk storage — TIDAK duplikasi UI storage di sini (lihat #17).

## Interaksi Wajib (semua halaman)

- Page transition: fade + slight scale (Framer Motion)
- Hero animation: shared layout transition artwork dari list → detail (pakai `layoutId` Framer Motion)
- Pull-to-refresh di Home (opsional, nice-to-have)
- Skeleton loading state untuk semua list yang fetch data (jangan blank/spinner polos)
