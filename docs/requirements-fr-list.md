# Functional Requirements (FR-01 – FR-38)

Daftar lengkap dari sesi requirement analysis awal (STEP 1) — sebelumnya
cuma pernah dibahas di chat perencanaan, tidak pernah tersimpan sebagai
file. Disimpan di sini permanen supaya jadi acuan cross-check yang bisa
diverifikasi ulang kapan saja (dipakai pertama kali di
`docs/audit-report-2026-08-05.md`, Bagian F).

- FR-01: User login via Google OAuth atau email/password
- FR-02: JWT access token + refresh token, auto-refresh tanpa re-login
- FR-03: Owner bisa invite Member (limited access) — tidak ada self-registration publik
- FR-04: Stream audio dengan HTTP Range support (pause/resume/seek/fast-forward/rewind)
- FR-05: Mini player persist di semua halaman, Full player dengan artwork besar
- FR-06: Queue management (add, reorder via drag, remove, clear)
- FR-07: Shuffle & Repeat (single/all)
- FR-08: Sleep timer, playback speed control
- FR-09: Continue Listening — resume dari posisi terakhir tiap device
- FR-10: History otomatis tercatat tiap play, Recently Played di Home
- FR-11: Playlist pribadi (create, edit, reorder song, delete)
- FR-12: Favorite (song, album, artist, playlist)
- FR-13: Artist detail (biography, popular songs, albums, related artist)
- FR-14: Album detail (tracklist, play/shuffle seluruh album)
- FR-15: Full-text search: song, artist, album, genre, lyrics, composer, producer, mood, language, year, playlist
- FR-16: Fuzzy search + typo tolerance + partial word match
- FR-17: Autocomplete/suggestion saat mengetik
- FR-18: Response time search <100ms
- FR-19: Lyrics auto-scroll dengan highlight baris aktif
- FR-20: Tap/klik baris lirik → seek audio ke posisi itu
- FR-21: Lyrics fullscreen mode dengan background blur dari artwork
- FR-22: Lyrics di-cache offline untuk lagu yang sudah didownload
- FR-23: Manual upload lagu (single/bulk) via UI
- FR-24: Ingest otomatis dari provider legal terautentikasi (Bandcamp purchases, cloud storage sync)
- FR-25: Pipeline otomatis: verify → upload Drive pool → waveform → metadata → lyrics → artwork → index search
- FR-26: Status tracking per ingest job (pending/running/completed/failed) di admin panel
- FR-27: Lagu gagal auto-ingest ditandai "Needs manual upload"
- FR-28: Support multiple Google Drive account sebagai satu storage pool
- FR-29: Auto pilih drive terbaik saat upload (quota-aware routing)
- FR-30: Health check per drive, auto-failover kalau drive down/rate-limited
- FR-31: Auto-rebalance/migration antar drive kalau ada yang hampir penuh
- FR-32: Playback state (posisi, queue) sync antar device real-time
- FR-33: PWA installable (Android/iPhone/Desktop), offline cache lagu terdownload
- FR-34: Admin Dashboard (total song, storage, drive, job stats)
- FR-35: Admin Drive manager (add/remove drive, quota, health)
- FR-36: Admin Ingest job monitor (pending/running/failed, retry)
- FR-37: Admin Lyrics provider priority config + health
- FR-38: Admin Analytics (top played, storage growth, ingest trend)
