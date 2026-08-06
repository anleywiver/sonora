# UI Implementation Spec — Literal, Wajib Diikuti Persis

Dokumen ini adalah level detail PALING RENDAH sebelum kode — setiap nilai
sudah final, tidak ada ruang untuk "kira-kira" atau improvisasi. Kalau ada
halaman yang tidak cocok dengan spec ini setelah diimplementasi, itu BUG,
harus diperbaiki sebelum halaman ditandai selesai.

Semua warna, radius, spacing di bawah SAMA PERSIS dengan `tailwind.config.ts`
yang sudah ada di project — dokumen ini menjelaskan CARA PAKAI-nya per
komponen, bukan mendefinisikan ulang.

---

## 0. APP SHELL — WAJIB ada di root layout, ini akar dari banyak bug tampilan

File: `apps/frontend/src/app/(main)/layout.tsx` dan `apps/admin/src/app/layout.tsx`

```tsx
export default function MainLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-background flex justify-center">
      <div className="w-full max-w-[430px] min-h-screen relative flex flex-col bg-background">
        <main className="flex-1 flex flex-col overflow-y-auto pb-24">
          {children}
        </main>
        <BottomNav />
        <MiniPlayer /> {/* render kondisional, cuma muncul kalau ada lagu aktif */}
      </div>
    </div>
  );
}
```

Checklist verifikasi (screenshot dan cek manual satu-satu):
- [ ] Di layar desktop lebar (>430px), ada area kosong warna `bg-background` di kiri DAN kanan konten — ini BENAR, bukan bug
- [ ] Konten tidak pernah melebar lebih dari 430px
- [ ] `pb-24` di `<main>` — WAJIB, supaya konten paling bawah tidak ketutup BottomNav+MiniPlayer yang fixed

---

## 1. ATOMS

### 1.1 Button — 3 varian, tidak ada varian lain

```tsx
// Primary — dipakai MAKSIMAL 1x per layar
<button className="bg-primary hover:bg-hover text-white rounded-2xl px-6 py-3 font-semibold text-sm transition-colors">
  {label}
</button>

// Secondary
<button className="bg-accent/10 hover:bg-accent/15 text-accent border border-accent/30 rounded-2xl px-6 py-3 font-semibold text-sm transition-colors">
  {label}
</button>

// Ghost
<button className="bg-transparent hover:bg-white/5 text-text-secondary border border-white/[0.06] rounded-2xl px-6 py-3 font-medium text-sm transition-colors">
  {label}
</button>
```

### 1.2 IconButton (tombol cuma icon, misal like/share/dots)

```tsx
<button className="w-9 h-9 flex items-center justify-center rounded-full hover:bg-white/5 transition-colors">
  <Icon className="w-5 h-5 text-text-secondary" />
</button>
```

### 1.3 Input (text field)

```tsx
<input className="w-full bg-white/5 border border-white/[0.06] rounded-2xl px-4 py-3 text-white text-sm placeholder:text-text-secondary focus:border-accent/50 focus:outline-none transition-colors" />
```

### 1.4 Badge (status)

```tsx
const badgeVariants = {
  success: "bg-[#4ADE80]/15 text-[#4ADE80]",
  warning: "bg-[#FACC15]/15 text-[#FACC15]",
  error:   "bg-[#F87171]/15 text-[#F87171]",
  info:    "bg-accent/15 text-accent",
};

<span className={`rounded-full px-3 py-1 text-[11px] font-semibold ${badgeVariants[variant]}`}>
  {label}
</span>
```

### 1.5 Avatar

```tsx
<img className="w-9 h-9 rounded-full object-cover" src={avatarUrl} />
// Fallback tanpa foto:
<div className="w-9 h-9 rounded-full bg-gradient-to-br from-accent to-primary flex items-center justify-center text-white text-xs font-bold">
  {initial}
</div>
```

### 1.6 Skeleton Loading

```tsx
<div className="bg-white/[0.06] rounded-xl animate-pulse" style={{ width, height }} />
```

---

## 2. MOLECULES

### 2.1 Card lagu/album/playlist (grid horizontal scroll)

```tsx
<div className="flex-shrink-0 w-[120px] bg-white/5 border border-white/[0.06] rounded-[18px] p-3.5 backdrop-blur-md">
  <div className="w-full aspect-square rounded-xl bg-gradient-to-br from-accent to-primary mb-2" />
  <p className="text-white text-[11px] font-semibold truncate">{title}</p>
  <p className="text-text-secondary text-[9px] truncate mt-0.5">{subtitle}</p>
</div>
```

### 2.2 SongRow

```tsx
<div className="flex items-center gap-2.5 py-2">
  <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
  <div className="flex-1 min-w-0">
    <p className="text-white text-xs font-semibold truncate">{title}</p>
    <p className="text-text-secondary text-[10px] truncate">{artist}</p>
  </div>
  <HeartIcon className={`w-4 h-4 flex-shrink-0 ${isFavorite ? "text-hover fill-hover" : "text-text-secondary"}`} />
  <span className="text-text-secondary text-[10px] flex-shrink-0">{duration}</span>
</div>
```

### 2.3 SearchBar

```tsx
<div className="bg-white/5 border border-white/[0.06] rounded-2xl px-4 py-2.5 flex items-center gap-2">
  <SearchIcon className="w-4 h-4 text-text-secondary flex-shrink-0" />
  <input
    placeholder="Song, artist, album, lyrics..."
    className="flex-1 bg-transparent text-white text-xs placeholder:text-text-secondary focus:outline-none"
  />
</div>
```

### 2.4 EmptyState (WAJIB dipakai di semua halaman list yang bisa kosong)

```tsx
interface EmptyStateProps {
  icon: LucideIcon;
  message: string;
  ctaLabel?: string;
  onCtaClick?: () => void;
}

<div className="flex flex-col items-center justify-center flex-1 px-8 text-center min-h-[400px]">
  <div className="w-16 h-16 rounded-2xl bg-white/5 border border-white/[0.06] flex items-center justify-center mb-4">
    <Icon className="w-7 h-7 text-text-secondary" />
  </div>
  <p className="text-text-secondary text-sm mb-5 max-w-[240px]">{message}</p>
  {ctaLabel && (
    <button className="bg-primary hover:bg-hover text-white rounded-2xl px-6 py-3 font-semibold text-sm transition-colors">
      {ctaLabel}
    </button>
  )}
</div>
```

Teks per halaman (PAKAI PERSIS INI):

| Halaman | Icon | Message | CTA |
|---|---|---|---|
| Favorite | `Heart` | "Belum ada favorite. Mulai dengan menandai lagu yang kamu suka." | "Cari lagu" |
| Library (Songs tab) | `Music` | "Koleksi lagu kamu masih kosong. Upload lagu pertamamu sekarang." | "Upload lagu" |
| Downloads | `Download` | "Belum ada lagu yang didownload untuk offline." | "Jelajahi lagu" |
| Search (hasil kosong) | `SearchX` | "Tidak ada hasil untuk pencarian ini." | (tanpa CTA) |
| Home (belum ada history) | `Sparkles` | "Belum ada lagu untuk ditampilkan. Mulai dengan mencari sesuatu." | "Cari lagu" |

### 2.5 Bottom Sheet

```tsx
<div className="fixed inset-0 bg-black/50 z-40" onClick={onClose} />
<div className="fixed bottom-0 inset-x-0 mx-auto max-w-[430px] bg-card border-t border-white/[0.06] rounded-t-3xl p-5 z-50">
  <div className="w-9 h-1 bg-white/20 rounded-full mx-auto mb-4" />
  {children}
</div>
```

### 2.6 Modal (dialog center)

```tsx
<div className="fixed inset-0 bg-black/50 z-40 flex items-center justify-center px-6" onClick={onClose}>
  <div className="w-full max-w-[380px] bg-card border border-white/[0.06] rounded-[20px] p-5" onClick={e => e.stopPropagation()}>
    {children}
  </div>
</div>
```

### 2.7 Toast

```tsx
<div className="fixed bottom-24 inset-x-0 mx-auto max-w-[400px] px-4 z-50">
  <div className="bg-white/[0.06] border border-white/[0.06] rounded-2xl px-4 py-3 flex items-center gap-2.5 backdrop-blur-md">
    <CheckIcon className="w-[18px] h-[18px] text-[#4ADE80] flex-shrink-0" />
    <span className="text-white text-xs font-medium">{message}</span>
  </div>
</div>
```

### 2.8 Tabs

```tsx
<div className="flex gap-1 bg-white/[0.04] rounded-[14px] p-1 w-fit">
  {tabs.map(tab => (
    <button
      key={tab.value}
      className={
        tab.value === active
          ? "bg-primary text-white px-4.5 py-2 rounded-[10px] text-xs font-semibold"
          : "text-text-secondary px-4.5 py-2 rounded-[10px] text-xs font-medium"
      }
    >
      {tab.label}
    </button>
  ))}
</div>
```

---

## 3. ORGANISMS

### 3.1 Bottom Navigation

```tsx
<nav className="fixed bottom-0 inset-x-0 mx-auto max-w-[430px] px-3 pb-2 z-30">
  <div className="bg-card/85 backdrop-blur-md rounded-[22px] px-5 py-2.5 flex justify-around items-center border border-white/[0.06]">
    {navItems.map(item => (
      <Link key={item.href} href={item.href} className="flex flex-col items-center gap-1">
        <item.icon className={`w-[19px] h-[19px] ${isActive ? "text-hover" : "text-text-secondary"}`} />
        <span className={`text-[10px] ${isActive ? "text-hover font-semibold" : "text-text-secondary font-normal"}`}>
          {item.label}
        </span>
      </Link>
    ))}
  </div>
</nav>
```

5 item persis urutan ini: Home (`house`), Search (`search`), Library (`library`), Favorite (`heart`), Settings (`settings`).

### 3.2 Mini Player

```tsx
<div className="fixed bottom-[76px] inset-x-0 mx-auto max-w-[430px] px-3 z-20">
  <div className="bg-white/[0.06] border border-white/[0.06] rounded-2xl backdrop-blur-md px-3 py-2 flex items-center gap-2.5">
    <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
    <div className="flex-1 min-w-0">
      <p className="text-white text-[11px] font-semibold truncate">{title}</p>
      <p className="text-text-secondary text-[9px] truncate">{artist}</p>
    </div>
    <button><TvIcon className="w-[15px] h-[15px] text-hover" /></button>
    <button><HeartIcon className="w-4 h-4 text-text-secondary" /></button>
    <button><PlayIcon className="w-4 h-4 text-white" /></button>
  </div>
</div>
```

`bottom-[76px]` supaya nempel PAS di atas Bottom Navigation.

### 3.3 Header halaman top-level (Home, Search, Library, Favorite, Settings)

```tsx
<header className="px-4 pt-5 pb-3 flex items-center justify-between">
  <h1 className="text-white text-lg font-bold">{pageTitle}</h1>
  {rightSlot}
</header>
```

### 3.4 Header halaman DETAIL dengan hero gradient (Song Detail, Playlist Detail, Album Detail)

```tsx
<div className="bg-gradient-to-b from-primary to-background pt-5 px-4 pb-6 rounded-b-[28px]">
  <div className="flex items-center justify-between mb-6">
    <button onClick={onBack}><ChevronDownIcon className="w-5 h-5 text-white" /></button>
    <button><MoreIcon className="w-[18px] h-[18px] text-white" /></button>
  </div>
</div>
```

---

## 4. SPEC PER HALAMAN (prioritas: paling sering diakses dulu)

### 4.1 Home (`/`)

```
Header (bukan judul statis, pakai greeting):
  Kiri: "Good evening" (text-secondary, 11px) + nama user (text-white, 18px, bold)
  Kanan: Avatar 34px

Section Continue Listening (render kondisional, cuma kalau ada data):
  Judul "Continue listening" (text-white, 12px, font-semibold, mb-2)
  Card horizontal: artwork 38px + judul + progress bar 3px height

Section Recently Played:
  Judul "Recently played"
  Horizontal scroll, Card artwork 76x76px, gap-2.5

Section Trending Now:
  Judul "Trending now"
  Horizontal scroll, Card 120px width (pakai molecule 2.1)

Section Quick Mix:
  Judul "Quick mix"
  Horizontal scroll, Card 110x56px gradient, label besar putih bold

EMPTY STATE (tidak ada history/data sama sekali):
  Ganti SEMUA section dengan 1 EmptyState (icon Sparkles, tabel 2.4)
```

### 4.2 Favorite (`/favorite`)

```
Header title="Favorite"
Tabs: Songs, Albums, Artists, Playlists (molecule 2.8)
Body: List SongRow (2.2), px-4 container, py-2 per row

EMPTY STATE: icon=Heart (tabel 2.4), WAJIB flex-1 supaya center vertikal
di sisa layar — BUKAN nempel di bawah header seperti bug yang terjadi
sebelumnya.
```

### 4.3 Library (`/library`)

```
Header title="Library"
Tabs: Songs, Albums, Artists, Playlists
Body: Grid 2 kolom (grid grid-cols-2 gap-3 px-4) untuk Albums/Artists/
Playlists, atau list SongRow untuk tab Songs

EMPTY STATE: icon=Music, CTA="Upload lagu"
```

### 4.4 Song Detail (`/song/:id`)

```
Header Detail dengan hero gradient (3.4)

Dalam hero:
  Artwork: w-[210px] h-[210px] rounded-2xl mx-auto mb-4
  Judul: text-white text-xl font-bold text-center mb-1
  Artist: text-white/65 text-sm text-center mb-1
  Album+tahun: text-white/45 text-[11px] text-center mb-5
  Action row (center, gap-5): Shuffle (white/70, 18px) — Play button
  (56px circle bg-white, icon 24px warna #050816) — Download (white/70, 18px)

Bottom action row (border-t border-white/10 pt-2.5):
  4 icon+label kecil: Like, Queue, Lyrics, Share
```

### 4.5 Now Playing (`/player/now-playing`)

```
Background (inline style, bukan Tailwind class biasa):
  style={{ background: 'radial-gradient(circle at 50% 20%, #1D4ED8 0%, #050816 65%)' }}

Header: chevron-down kiri, "PLAYING FROM PLAYLIST" tengah (white/60, 10px,
font-semibold, letter-spacing wide), dots kanan

Artwork: w-[190px] h-[190px] rounded-[18px] mx-auto mb-4.5

Info row: judul (white, 17px, bold) + artist (white/60, 12px) kiri,
Heart icon kanan (18px)

Seekbar: h-1 bg-white/15 rounded-full, progress bg-white rounded-full
Waktu: flex justify-between, white/50, 10px

Kontrol utama (flex justify-center items-center gap-5):
  Shuffle (18px) — Prev (white, 22px) — Play/Pause (56px circle bg-white,
  icon 22px warna #050816) — Next (white, 22px) — Repeat (white/50, 18px)

Bottom utility row (border-t border-white/10 pt-3.5):
  4 kolom: Sleep Timer, Speed (1.0x), Lyrics (warna hover #60A5FA), Queue
  — icon 16px + label 9px
```

---

## 5. ATURAN VERIFIKASI WAJIB

Setelah SETIAP halaman selesai diimplementasi:

1. Screenshot halaman itu
2. Buka Chrome DevTools → Inspect element pada 3 elemen acak (background utama, 1 card, 1 tombol)
3. Cek nilai computed CSS `background-color` / `border-radius` — cocokkan MANUAL ke hex/px di dokumen ini
4. Kalau ADA yang tidak cocok (misal border-radius keluar 8px padahal spec minta 18px), itu BUG — perbaiki dulu, jangan lanjut halaman berikutnya
5. Laporkan hasil verifikasi ke user PER HALAMAN, bukan cuma "sudah selesai semua" di akhir

---

## 6. PERBAIKAN KRITIS — hasil audit QA (WAJIB dikerjakan sebelum lanjut fitur baru)

### 6.1 Mini Player HARUS hilang saat Now Playing full-screen terbuka

Ini bug paling serius — Mini Player dan Now Playing tumpuk jadi 2 lapis.
Root cause: Mini Player dirender global di App Shell (layout.tsx), tidak
ada logic untuk sembunyikan dia saat user pindah ke rute full-screen player.

Perbaikan di App Shell (Bagian 0):

```tsx
'use client';
import { usePathname } from 'next/navigation';

const FULLSCREEN_PLAYER_ROUTES = ['/player/now-playing', '/player/lyrics', '/player/queue'];

export default function MainLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const isFullscreenPlayer = FULLSCREEN_PLAYER_ROUTES.includes(pathname);

  return (
    <div className="min-h-screen bg-background flex justify-center">
      <div className="w-full max-w-[430px] min-h-screen relative flex flex-col bg-background">
        <main className="flex-1 flex flex-col overflow-y-auto pb-24">
          {children}
        </main>
        {!isFullscreenPlayer && <BottomNav />}
        {!isFullscreenPlayer && <MiniPlayer />}
      </div>
    </div>
  );
}
```

### 6.2 Now Playing — lengkapi kontrol yang hilang

Halaman ini SEKARANG cuma punya artwork + 1 tombol play/pause. Tambahkan
SEMUA elemen dari spec 4.5 yang belum ada:
- Kontrol utama lengkap: Shuffle — Prev — Play/Pause — Next — Repeat
  (bukan cuma Play/Pause sendirian)
- Bottom utility row 4 kolom: Sleep Timer, Speed (1.0x), **Lyrics**, Queue
  — tombol Lyrics ini yang navigate ke /player/lyrics (fullscreen lyrics
  view), INI YANG BIKIN FITUR LYRICS "TIDAK ADA" — tombolnya memang belum
  dibuat sama sekali

### 6.3 Settings — bangun sekarang, jangan ditunda ke "sprint lanjutan"

Halaman ini WAJIB langsung dibangun lengkap (bukan placeholder), struktur:

```tsx
<Header title="Settings" />

<div className="px-4 space-y-6">
  {/* Group: Appearance */}
  <div>
    <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Appearance</p>
    <div className="flex items-center justify-between py-2.5 border-b border-white/[0.06]">
      <span className="text-white text-xs font-medium">Theme</span>
      <span className="text-text-secondary text-[11px] flex items-center gap-1">Dark <ChevronRightIcon className="w-3.5 h-3.5" /></span>
    </div>
  </div>

  {/* Group: Storage */}
  <div>
    <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Storage</p>
    <div className="py-2.5">
      <div className="flex justify-between mb-1.5">
        <span className="text-white text-xs font-medium">Storage used</span>
        <span className="text-text-secondary text-[11px]">{used} / {total}</span>
      </div>
      <div className="h-[5px] bg-white/[0.08] rounded-full">
        <div className="h-full bg-accent rounded-full" style={{ width: `${percent}%` }} />
      </div>
    </div>
    <div className="flex items-center justify-between py-2.5 border-b border-white/[0.06]">
      <div>
        <p className="text-white text-xs font-medium">Connected Google Drive</p>
        <p className="text-text-secondary text-[10px] mt-0.5">{count} accounts linked</p>
      </div>
      <ChevronRightIcon className="w-3.5 h-3.5 text-text-secondary" />
    </div>
    <div className="flex items-center justify-between py-2.5 border-b border-white/[0.06]">
      <span className="text-white text-xs font-medium">Clear cache</span>
      <span className="text-text-secondary text-[11px]">{cacheSize}</span>
    </div>
  </div>

  {/* Group: Playback */}
  <div>
    <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Playback</p>
    <div className="flex items-center justify-between py-2.5 border-b border-white/[0.06]">
      <span className="text-white text-xs font-medium">Lyrics source</span>
      <span className="text-text-secondary text-[11px] flex items-center gap-1">Auto <ChevronRightIcon className="w-3.5 h-3.5" /></span>
    </div>
  </div>

  {/* Group: About */}
  <div className="flex items-center justify-between py-2.5">
    <span className="text-white text-xs font-medium">About</span>
    <span className="text-text-secondary text-[11px]">v1.0.0</span>
  </div>
</div>
```

### 6.4 Home — perbaiki section Trending (salah komponen) + tambah section yang hilang

Trending SAAT INI pakai SongRow (list vertikal) — SALAH. Harus pakai Card
molecule (2.1) dalam horizontal scroll container:

```tsx
<div>
  <p className="text-white text-xs font-semibold mb-2 px-4">Trending now</p>
  <div className="flex gap-2.5 overflow-x-auto px-4 pb-1">
    {trendingSongs.map(song => <Card key={song.id} {...song} />)}
  </div>
</div>
```

Tambahkan section yang belum ada sama sekali: Continue Listening, Recently
Played, Quick Mix (lihat spec lengkap di 4.1) — kalau data belum tersedia
untuk section tertentu (misal belum ada history), section itu boleh
disembunyikan sementara, JANGAN ditampilkan kosong tanpa isi.

Greeting: ganti "Halo 👋" jadi dinamis:
```tsx
const hour = new Date().getHours();
const greeting = hour < 11 ? "Good morning" : hour < 15 ? "Good afternoon" : "Good evening";
// tampilkan: {greeting} + nama user dari /auth/me, BUKAN teks statis
```

### 6.5 Admin — Songs table: tambah thumbnail, detail, dan Add

```tsx
<table>
  <thead>
    <tr>
      <th></th> {/* kolom thumbnail, tanpa label */}
      <th>Judul</th>
      <th>Artist</th>
      <th>Album</th>
      <th>Durasi</th>
      <th>Storage</th>
      <th>Ditambahkan</th>
      <th>Action</th>
    </tr>
  </thead>
  <tbody>
    {songs.map(song => (
      <tr
        key={song.id}
        className="cursor-pointer hover:bg-white/[0.03] transition-colors"
        onClick={() => router.push(`/songs/${song.id}`)}
      >
        <td className="w-10">
          <div className="w-8 h-8 rounded-md bg-gradient-to-br from-accent to-primary" />
          {/* atau <img> kalau artwork_url ada */}
        </td>
        <td>{song.title}</td>
        {/* ...kolom lain sama seperti sekarang */}
        <td onClick={e => e.stopPropagation()}>
          {/* tombol edit/delete tetap ada, stopPropagation supaya klik
              tombol tidak ikut trigger navigate ke detail */}
        </td>
      </tr>
    ))}
  </tbody>
</table>
```

Buat halaman baru `apps/admin/src/app/songs/[id]/page.tsx` — detail lagu:
artwork besar, semua metadata (title, artist, album, genre, tags, credits,
storage location, checksum), form edit inline, tombol delete dengan
konfirmasi.

Tombol "Add Song" (manual, dari admin) — taruh di kanan atas halaman
Songs, di sebelah search bar, buka modal upload sama seperti flow user
biasa (reuse endpoint POST /ingest/upload).

---

## 7. SPEC LENGKAP — SEMUA HALAMAN YANG BELUM TERCOVER

Ini menutup SISA halaman dari 25 total (di luar yang sudah di Bagian 4 & 6).
Semua tetap wajib pakai komponen Bagian 1-3, jangan bikin versi baru.

### 7.1 Splash Screen (`/splash`, auto-redirect)

```tsx
<div className="min-h-screen bg-background flex flex-col items-center justify-center relative overflow-hidden">
  <div className="absolute top-[30%] left-1/2 -translate-x-1/2 w-[260px] h-[260px] rounded-full animate-pulse"
       style={{ background: 'radial-gradient(circle, #1D4ED8 0%, rgba(29,78,216,0) 70%)' }} />
  <div className="relative z-10 text-center">
    <div className="w-16 h-16 rounded-[20px] bg-gradient-to-br from-accent to-primary mx-auto mb-4 flex items-center justify-center">
      <HeadphonesIcon className="w-8 h-8 text-white" />
    </div>
    <h1 className="text-white text-xl font-bold tracking-wide">Sonora</h1>
    <p className="text-text-secondary text-[11px] uppercase tracking-widest mt-1">Your sound, your space</p>
  </div>
  <div className="absolute bottom-14 flex gap-1.5">
    {[0,1,2].map(i => (
      <span key={i} className="w-1.5 h-1.5 bg-hover rounded-full animate-bounce" style={{ animationDelay: `${i*0.2}s` }} />
    ))}
  </div>
</div>
```
Logic: `useEffect` cek token valid di localStorage/cookie → kalau ada, redirect `/` (Home). Kalau tidak ada, tunggu 1.5s lalu redirect `/login`.

### 7.2 Login (User) — `/login`

```tsx
<div className="min-h-screen bg-background flex flex-col justify-center px-6 py-9">
  <div className="text-center mb-8">
    <div className="w-12 h-12 rounded-[14px] bg-gradient-to-br from-accent to-primary mx-auto mb-3.5 flex items-center justify-center">
      <HeadphonesIcon className="w-6 h-6 text-white" />
    </div>
    <h1 className="text-white text-xl font-bold mb-1.5">Welcome back</h1>
    <p className="text-text-secondary text-xs">Sign in to keep your music with you</p>
  </div>

  {/* Fetch GET /auth/config saat mount, render tombol ini HANYA kalau google_oauth_enabled true */}
  {googleOAuthEnabled && (
    <button className="bg-white rounded-2xl px-4 py-3 flex items-center justify-center gap-2.5 mb-4">
      <GoogleIcon className="w-[18px] h-[18px]" />
      <span className="text-[#050816] text-[13px] font-semibold">Continue with Google</span>
    </button>
  )}

  <input placeholder="Username" className="w-full bg-white/5 border border-white/[0.06] rounded-2xl px-4 py-3 text-white text-sm placeholder:text-text-secondary mb-2.5" />
  <input placeholder="Password" type="password" className="w-full bg-white/5 border border-white/[0.06] rounded-2xl px-4 py-3 text-white text-sm placeholder:text-text-secondary mb-4" />
  <button className="bg-primary hover:bg-hover text-white rounded-2xl px-6 py-3 font-semibold text-sm w-full mb-3.5">Sign in</button>

  <p className="text-text-secondary text-[11px] text-center">
    Belum punya akses? <a href={`https://wa.me/${process.env.NEXT_PUBLIC_OWNER_WHATSAPP}?text=Halo...`} target="_blank" className="text-hover font-semibold">Request akses</a>
  </p>
</div>
```

### 7.3 Admin Login — `/login` (apps/admin)

Sama seperti 7.2, TAPI: hilangkan link "Request akses" (tidak relevan untuk admin), tombol Google (kalau enabled) langsung di atas form email+password (bukan username).

### 7.4 Search — `/search`

```tsx
<Header title="Search" />
<div className="px-4">
  <SearchBar />

  <div className="mt-5">
    <div className="flex justify-between items-center mb-2.5">
      <p className="text-white text-xs font-semibold">Recent searches</p>
      <button className="text-hover text-[10px]">Clear</button>
    </div>
    {recentSearches.map(term => (
      <div key={term} className="flex items-center gap-2 py-1.5">
        <ClockIcon className="w-3.5 h-3.5 text-text-secondary" />
        <span className="text-white text-xs">{term}</span>
      </div>
    ))}
  </div>

  <div className="mt-4.5">
    <p className="text-white text-xs font-semibold mb-2.5">Trending searches</p>
    <div className="flex flex-wrap gap-2">
      {trending.map(t => (
        <span key={t} className="bg-accent/10 text-accent px-3 py-1.5 rounded-full text-[11px] font-medium">{t}</span>
      ))}
    </div>
  </div>

  <div className="mt-4.5">
    <p className="text-white text-xs font-semibold mb-2.5">Browse genre</p>
    <div className="grid grid-cols-2 gap-2.5">
      {genres.map(g => (
        <div key={g.name} className="h-14 rounded-xl flex items-center px-3" style={{ background: g.gradient }}>
          <span className="text-white text-[11px] font-bold">{g.name}</span>
        </div>
      ))}
    </div>
  </div>
</div>
```

### 7.5 Search Result — `/search/results?q=`

```tsx
<div className="px-4 pt-5">
  <SearchBar defaultValue={query} />
  <div className="mt-4 mb-4.5"><Tabs tabs={['All','Songs','Artists','Lyrics']} /></div>

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Top result</p>
  <div className="flex items-center gap-2.5 mb-4.5">
    <div className="w-[52px] h-[52px] rounded-full bg-gradient-to-br from-hover to-primary" />
    <div><p className="text-white text-[13px] font-bold">{name}</p><p className="text-text-secondary text-[10px] mt-0.5">{type} · {meta}</p></div>
  </div>

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Songs</p>
  {songs.map(s => <SongRow key={s.id} {...s} />)}

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2 mt-4.5">Lyrics match</p>
  <div className="bg-white/[0.04] rounded-[10px] p-2.5">
    <p className="text-white text-[11px] font-semibold mb-0.5">"{lyricSnippet}..."</p>
    <p className="text-text-secondary text-[10px]">{song} · {artist}</p>
  </div>
</div>
```
EMPTY STATE (semua tab kosong): icon `SearchX`, message "Tidak ada hasil untuk pencarian ini.", tanpa CTA.

### 7.6 Lyrics fullscreen — `/player/lyrics` (INI YANG BELUM PERNAH DIBUAT — prioritas tinggi)

```tsx
<div className="min-h-screen bg-gradient-to-b from-card to-background flex flex-col px-5.5 pt-5">
  <div className="flex justify-between items-center mb-2">
    <ChevronDownIcon className="w-5 h-5 text-white" onClick={onClose} />
    <div className="flex items-center gap-1.5">
      <div className="w-[22px] h-[22px] rounded-md bg-gradient-to-br from-hover to-primary" />
      <span className="text-white/70 text-[10px] font-semibold">{songTitle}</span>
    </div>
    <MoreIcon className="w-[18px] h-[18px] text-white" />
  </div>

  <div className="flex-1 flex flex-col justify-center gap-4">
    {lyricsLines.map((line, i) => {
      const distance = Math.abs(i - activeLineIndex);
      const opacity = distance === 0 ? 1 : distance === 1 ? 0.45 : 0.3;
      const isActive = i === activeLineIndex;
      return (
        <p
          key={i}
          onClick={() => seekTo(line.timestampMs)}
          className={isActive
            ? "text-white text-lg font-bold bg-accent/15 rounded-xl px-2.5 -mx-2.5 cursor-pointer"
            : "text-white/45 text-sm font-semibold cursor-pointer"}
          style={{ opacity: isActive ? 1 : opacity }}
        >
          {line.text}
        </p>
      );
    })}
  </div>

  <div className="flex justify-between items-center pt-3.5 border-t border-white/[0.08] pb-5">
    <span className="text-white/50 text-[10px]">Tap a line to jump</span>
    <div className="flex items-center gap-1.5">
      <PlayIcon className="w-3.5 h-3.5 text-white" />
      <span className="text-white/50 text-[10px]">Auto-scroll</span>
    </div>
  </div>
</div>
```
Dipanggil dari tombol "Lyrics" di Now Playing (Bagian 6.2). Auto-scroll: pakai `scrollIntoView({ behavior: 'smooth', block: 'center' })` pada baris aktif setiap `activeLineIndex` berubah. Kalau lagu tidak punya lirik tersinkron, tampilkan EmptyState (icon `MicOff`, message "Lirik tidak tersedia untuk lagu ini.").

### 7.7 Queue — `/player/queue`

```tsx
<Header title="Queue" rightSlot={<button className="text-[#F87171] text-[11px] font-semibold">Clear queue</button>} />
<div className="px-4">
  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2">Now playing</p>
  <div className="bg-accent/10 border border-accent/25 rounded-xl p-2.5 flex items-center gap-2.5 mb-4.5">
    <div className="w-[38px] h-[38px] rounded-lg bg-gradient-to-br from-hover to-primary flex-shrink-0" />
    <div className="flex-1 min-w-0"><p className="text-white text-xs font-bold truncate">{title}</p><p className="text-[#93C5FD] text-[10px]">{artist} · playing</p></div>
    <PauseIcon className="w-4 h-4 text-hover" />
  </div>

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2.5">Next up</p>
  {queueItems.map(item => (
    <div key={item.id} className="flex items-center gap-2.5 py-2">
      <GripVerticalIcon className="w-3.5 h-3.5 text-text-secondary flex-shrink-0 cursor-grab" />
      <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
      <div className="flex-1 min-w-0"><p className="text-white text-xs font-semibold truncate">{item.title}</p><p className="text-text-secondary text-[10px] truncate">{item.artist}</p></div>
      <XIcon className="w-3.5 h-3.5 text-text-secondary flex-shrink-0" onClick={() => remove(item.id)} />
    </div>
  ))}
</div>
```
Drag reorder: pakai `@dnd-kit/sortable`, tiap drop → `PATCH /queue/:id` dengan posisi fractional baru.
EMPTY STATE (queue kosong): icon `ListMusic`, message "Queue kosong. Tambahkan lagu untuk diputar berikutnya.".

### 7.8 Playlist Detail — `/playlist/:id`

Header Detail hero (3.4), lalu:
```tsx
<div className="w-[150px] h-[150px] rounded-2xl bg-gradient-to-br from-hover to-primary mx-auto mb-3.5" />
<h1 className="text-white text-lg font-bold text-center mb-1.5">{name}</h1>
<p className="text-white/60 text-[11px] text-center mb-1">{description}</p>
<p className="text-white/40 text-[10px] text-center mb-4">{songCount} songs · {duration}</p>
<div className="flex items-center gap-2.5 mb-3.5">
  <div className="w-11 h-11 rounded-full bg-white flex items-center justify-center"><PlayIcon className="w-5 h-5 text-background" /></div>
  <ShuffleIcon className="w-[18px] h-[18px] text-white" />
  <HeartIcon className="w-[18px] h-[18px] text-white" />
  <div className="flex-1" />
  <SearchIcon className="w-4 h-4 text-white" />
  <ArrowUpDownIcon className="w-4 h-4 text-white" />
</div>
{songs.map((s, i) => (
  <div key={s.id} className="flex items-center gap-2.5 py-1.5">
    <span className="text-white/40 text-[11px] w-3.5">{i+1}</span>
    <div className="w-8.5 h-8.5 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
    <div className="flex-1 min-w-0"><p className="text-white text-xs font-semibold truncate">{s.title}</p><p className="text-white/50 text-[10px] truncate">{s.artist}</p></div>
    <span className="text-white/40 text-[10px]">{s.duration}</span>
  </div>
))}
```
EMPTY STATE: icon `Music`, message "Playlist ini masih kosong. Tambahkan lagu dari halaman manapun.", tanpa CTA.

### 7.9 Artist Detail — `/artist/:id`

```tsx
<div className="h-[180px] bg-gradient-to-br from-primary to-background relative">
  <ChevronLeftIcon className="w-5 h-5 text-white absolute top-3.5 left-3.5" onClick={onBack} />
  <div className="absolute -bottom-[30px] left-4.5 w-[76px] h-[76px] rounded-full bg-gradient-to-br from-hover to-primary border-[3px] border-background" />
</div>
<div className="pt-10 px-4.5">
  <h1 className="text-white text-lg font-bold mb-0.5">{name}</h1>
  <p className="text-text-secondary text-[11px] mb-3.5">{monthlyListeners} monthly listeners</p>
  <div className="flex items-center gap-2.5 mb-4.5">
    <div className="w-10 h-10 rounded-full bg-primary flex items-center justify-center"><PlayIcon className="w-[18px] h-[18px] text-white" /></div>
    <button className="border border-white/[0.06] text-white text-[11px] font-semibold px-4 py-2 rounded-full">Follow</button>
  </div>

  <p className="text-white text-xs font-semibold mb-2.5">Popular</p>
  {popularSongs.map(s => (
    <div key={s.id} className="flex items-center gap-2.5 mb-2.5">
      <div className="w-8.5 h-8.5 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
      <p className="text-white text-xs font-medium flex-1 truncate">{s.title}</p>
      <span className="text-text-secondary text-[10px]">{s.duration}</span>
    </div>
  ))}

  <p className="text-white text-xs font-semibold mb-2.5 mt-4">Related artists</p>
  <div className="flex gap-3.5 overflow-x-auto">
    {related.map(a => (
      <div key={a.id} className="text-center flex-shrink-0">
        <div className="w-[50px] h-[50px] rounded-full bg-gradient-to-br from-hover to-primary mx-auto mb-1.5" />
        <p className="text-white text-[9px]">{a.name}</p>
      </div>
    ))}
  </div>

  <p className="text-white text-xs font-semibold mb-2 mt-4">Biography</p>
  <p className="text-text-secondary text-xs leading-relaxed">{biography}</p>
</div>
```

### 7.10 Album Detail — `/album/:id`

Struktur sama persis dengan Playlist Detail (7.8), bedanya:
- Subtitle row tambahan: `<p className="text-white/40 text-[10px] text-center mb-4">{year} · {genre} · {trackCount} tracks</p>`
- Tracklist pakai SongRow (2.2) LENGKAP dengan heart icon per track (Playlist Detail tidak perlu heart)

### 7.11 Downloads — `/downloads`

```tsx
<Header title="Downloads" />
<div className="px-4">
  <p className="text-text-secondary text-[11px] mb-4">{count} songs · {totalSize} used offline</p>
  <div className="mb-4.5"><Tabs tabs={['All','Downloading','Failed']} /></div>

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2.5">In progress</p>
  {downloading.map(d => (
    <div key={d.id} className="flex items-center gap-2.5 mb-3">
      <div className="w-9.5 h-9.5 rounded-lg bg-gradient-to-br from-hover to-primary flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-white text-xs font-semibold truncate mb-1">{d.title}</p>
        <div className="h-[3px] bg-white/10 rounded-full"><div className="h-full bg-accent rounded-full" style={{width: `${d.percent}%`}} /></div>
      </div>
      <span className="text-[#93C5FD] text-[10px] flex-shrink-0">{d.percent}%</span>
    </div>
  ))}

  {failed.map(d => (
    <div key={d.id} className="flex items-center gap-2.5 mb-3">
      <div className="w-9.5 h-9.5 rounded-lg bg-gradient-to-br from-[#EF4444] to-card flex-shrink-0" />
      <div className="flex-1 min-w-0">
        <p className="text-white text-xs font-semibold truncate mb-0.5">{d.title}</p>
        <p className="text-[#F87171] text-[10px]">Failed · tap to retry</p>
      </div>
      <RefreshCwIcon className="w-[15px] h-[15px] text-[#F87171] flex-shrink-0" />
    </div>
  ))}

  <p className="text-text-secondary text-[10px] font-semibold uppercase tracking-wide mb-2.5">Available offline</p>
  {downloaded.map(d => (
    <div key={d.id} className="flex items-center gap-2.5 mb-3">
      <div className="w-9.5 h-9.5 rounded-lg bg-gradient-to-br from-accent to-primary flex-shrink-0" />
      <div className="flex-1 min-w-0"><p className="text-white text-xs font-semibold truncate">{d.title}</p><p className="text-text-secondary text-[10px] truncate">{d.artist} · {d.fileSize}</p></div>
      <CheckCircle2Icon className="w-4 h-4 text-[#4ADE80] flex-shrink-0" />
    </div>
  ))}
</div>
```
EMPTY STATE: icon `Download` (tabel 2.4, sudah tercatat sebelumnya).

### 7.12 Profile — `/profile`

```tsx
<Header title="Profile" />
<div className="flex flex-col items-center px-4 pt-4">
  <div className="relative mb-4">
    <div className="w-[88px] h-[88px] rounded-full bg-gradient-to-br from-accent to-primary flex items-center justify-center text-white text-2xl font-bold">
      {initial}
    </div>
    <button className="absolute bottom-0 right-0 w-7 h-7 rounded-full bg-primary flex items-center justify-center border-2 border-background">
      <CameraIcon className="w-3.5 h-3.5 text-white" />
    </button>
  </div>
  <input value={name} className="bg-white/5 border border-white/[0.06] rounded-2xl px-4 py-2.5 text-white text-sm text-center mb-2.5 w-full max-w-[240px]" />
  <p className="text-text-secondary text-xs mb-3">{email}</p>
  <span className="bg-accent/15 text-accent rounded-full px-3 py-1 text-[11px] font-semibold mb-6">{role}</span>
  <button className="bg-transparent hover:bg-white/5 text-text-secondary border border-white/[0.06] rounded-2xl px-6 py-3 font-medium text-sm w-full">Sign out</button>
</div>
```

---

## 8. ADMIN — SEMUA HALAMAN (apps/admin)

### 8.0 Admin Shell (sidebar + content, dipakai SEMUA halaman admin)

```tsx
<div className="min-h-screen bg-background flex">
  <aside className="w-[220px] flex-shrink-0 bg-[#0A0F1F] border-r border-white/[0.06] p-5">
    <div className="flex items-center gap-2 mb-7">
      <div className="w-7 h-7 rounded-lg bg-gradient-to-br from-accent to-primary" />
      <span className="text-white text-sm font-bold">Sonora Admin</span>
    </div>
    <nav className="flex flex-col gap-0.5">
      {navItems.map(item => (
        <Link key={item.href} href={item.href}
          className={isActive
            ? "flex items-center gap-2 px-2.5 py-2 rounded-lg bg-accent/[0.12] text-hover text-[11px] font-semibold"
            : "flex items-center gap-2 px-2.5 py-2 rounded-lg text-text-secondary text-[11px]"}>
          <item.icon className="w-[14px] h-[14px]" />
          {item.label}
        </Link>
      ))}
    </nav>
  </aside>
  <main className="flex-1 p-6 overflow-y-auto">{children}</main>
</div>
```
9 nav item urutan ini: Dashboard, Drive Manager, Ingest Sources, Lyrics Source, Job Queue, Analytics, Users, Songs, Settings.

### 8.1 Dashboard

```tsx
<h1 className="text-white text-[15px] font-bold mb-4">Dashboard overview</h1>
<div className="grid grid-cols-4 gap-2.5 mb-5">
  {stats.map(s => (
    <div key={s.label} className="bg-card rounded-xl p-3">
      <p className="text-text-secondary text-[10px] mb-1.5">{s.label}</p>
      <p className="text-white text-lg font-bold">{s.value}</p>
    </div>
  ))}
</div>
<div className="grid grid-cols-[1.3fr_1fr] gap-3.5 mb-4.5">
  <div className="bg-card rounded-2xl p-4">
    <p className="text-white text-xs font-semibold mb-3">Storage distribution</p>
    {drives.map(d => (
      <div key={d.id} className="mb-2.5">
        <div className="flex justify-between mb-1"><span className="text-text-secondary text-[10px]">{d.label}</span><span className="text-text-secondary text-[10px]">{d.used} / {d.total}</span></div>
        <div className="h-[5px] bg-white/[0.08] rounded-full"><div className={`h-full rounded-full ${d.percent > 90 ? "bg-[#F87171]" : "bg-accent"}`} style={{width: `${d.percent}%`}} /></div>
      </div>
    ))}
  </div>
  <div className="bg-card rounded-2xl p-4">
    <p className="text-white text-xs font-semibold mb-3">Background jobs</p>
    {jobStats.map(j => (
      <div key={j.label} className="flex justify-between mb-2.5"><span className="text-text-secondary text-[11px]">{j.label}</span><span className={`text-[11px] font-semibold`} style={{color: j.color}}>{j.count}</span></div>
    ))}
  </div>
</div>
<div className="bg-card rounded-2xl p-4">
  <p className="text-white text-xs font-semibold mb-3">Top played this week</p>
  {topPlayed.map((t, i) => (
    <div key={t.id} className="flex items-center gap-2.5 mb-2"><span className="text-text-secondary text-[10px] w-3.5">{i+1}</span><span className="text-white text-[11px] flex-1">{t.title} — {t.artist}</span><span className="text-text-secondary text-[10px]">{t.plays} plays</span></div>
  ))}
</div>
```

### 8.2 Drive Manager

```tsx
<h1 className="text-white text-[15px] font-bold mb-1">Google Drive manager</h1>
<p className="text-text-secondary text-[11px] mb-3.5">Multiple drive support with dynamic allocation</p>
<div className="grid grid-cols-2 gap-2.5 mb-5.5">
  {drives.map(d => (
    <div key={d.id} className="bg-card rounded-xl p-3">
      <div className="flex justify-between items-center mb-2">
        <span className="text-white text-xs font-semibold">{d.label}</span>
        <span className={d.percent > 90 ? "bg-[#F87171]/15 text-[#F87171] text-[9px] font-semibold px-2 py-0.5 rounded-lg" : "bg-[#4ADE80]/15 text-[#4ADE80] text-[9px] font-semibold px-2 py-0.5 rounded-lg"}>
          {d.percent > 90 ? "Near full" : "Healthy"}
        </span>
      </div>
      <div className="h-[5px] bg-white/[0.08] rounded-full mb-1.5"><div className={d.percent > 90 ? "h-full bg-[#F87171] rounded-full" : "h-full bg-accent rounded-full"} style={{width: `${d.percent}%`}} /></div>
      <p className="text-text-secondary text-[10px]">{d.used} used · {d.remaining} remaining</p>
    </div>
  ))}
</div>
<div className="flex gap-2">
  <button className="bg-primary text-white px-4 py-2 rounded-[10px] text-[11px] font-semibold">+ Add drive</button>
  <button className="border border-white/[0.06] text-text-secondary px-4 py-2 rounded-[10px] text-[11px]">Run health check</button>
</div>
```
"Add drive" buka Modal (2.6) berisi input Refresh Token (textarea) + Label — sesuai flow manual OAuth Playground yang kita pakai.

### 8.3 Ingest Sources (dulu "Crawler")

```tsx
<h1 className="text-white text-[15px] font-bold mb-3.5">Ingest sources</h1>
{sources.map(s => (
  <div key={s.type} className="bg-card rounded-xl p-3.5 mb-2.5 flex items-center justify-between">
    <div><p className="text-white text-xs font-semibold">{s.name}</p><p className="text-text-secondary text-[10px] mt-0.5">{s.description}</p></div>
    <span className={s.connected ? "bg-[#4ADE80]/15 text-[#4ADE80] text-[9px] font-semibold px-2.5 py-1 rounded-lg" : "bg-white/5 text-text-secondary text-[9px] font-semibold px-2.5 py-1 rounded-lg"}>
      {s.connected ? "Connected" : "Not connected"}
    </span>
  </div>
))}

{/* Filter Rules panel, cuma untuk Bandcamp & Cloud Sync (bukan Manual Upload) */}
<div className="bg-card rounded-2xl p-4 mt-4">
  <p className="text-white text-xs font-semibold mb-3">Filter rules — Bandcamp & Cloud Sync</p>
  <p className="text-text-secondary text-[10px] mb-3">Cuma berlaku untuk auto-ingest, TIDAK berlaku untuk manual upload.</p>
  <p className="text-text-secondary text-[10px] font-semibold mb-1.5">Genre allow-list</p>
  <div className="flex flex-wrap gap-1.5 mb-3">{genres.map(g => <span key={g} className="bg-accent/15 text-accent text-[10px] px-2.5 py-1 rounded-full">{g} ×</span>)}</div>
  <p className="text-text-secondary text-[10px] font-semibold mb-1.5">Year range</p>
  <div className="flex gap-2 items-center"><input className="w-20 bg-white/5 border border-white/[0.06] rounded-lg px-2.5 py-1.5 text-white text-xs" placeholder="From" /><span className="text-text-secondary text-xs">—</span><input className="w-20 bg-white/5 border border-white/[0.06] rounded-lg px-2.5 py-1.5 text-white text-xs" placeholder="To" /></div>
</div>
```

### 8.4 Lyrics Source

```tsx
<h1 className="text-white text-[15px] font-bold mb-1">Lyrics source</h1>
<p className="text-text-secondary text-[11px] mb-3.5">Priority order used when fetching lyrics</p>
<div className="bg-card rounded-2xl overflow-hidden">
  <div className="grid grid-cols-[0.3fr_1.4fr_0.8fr_0.8fr] px-3 py-2.5 border-b border-white/[0.06]">
    {['#','Provider','Health','Match rate'].map(h => <span key={h} className="text-text-secondary text-[9px] font-semibold uppercase">{h}</span>)}
  </div>
  {providers.map((p, i) => (
    <div key={p.id} className="grid grid-cols-[0.3fr_1.4fr_0.8fr_0.8fr] px-3 py-2.5 border-b border-white/[0.06] items-center">
      <span className="text-text-secondary text-[11px] flex items-center gap-1"><GripVerticalIcon className="w-3 h-3 cursor-grab" />{i+1}</span>
      <span className="text-white text-[11px] font-semibold">{p.name}</span>
      <span className={p.healthy ? "bg-[#4ADE80]/15 text-[#4ADE80] text-[9px] font-semibold px-2 py-0.5 rounded-lg w-fit" : "bg-[#F87171]/15 text-[#F87171] text-[9px] font-semibold px-2 py-0.5 rounded-lg w-fit"}>{p.healthStatus}</span>
      <span className="text-text-secondary text-[10px]">{p.matchRate}%</span>
    </div>
  ))}
</div>
```
Reorder via drag-drop → `PATCH /admin/lyrics-providers/:id` update priority.

### 8.5 Job Queue

```tsx
<h1 className="text-white text-[15px] font-bold mb-3.5">Job queue</h1>
<div className="bg-card rounded-2xl overflow-hidden">
  <div className="grid grid-cols-[1.5fr_0.8fr_0.8fr_1fr] px-3 py-2.5 border-b border-white/[0.06]">
    {['Job','Type','Status','Action'].map(h => <span key={h} className="text-text-secondary text-[9px] font-semibold uppercase">{h}</span>)}
  </div>
  {jobs.map(j => (
    <div key={j.id} className="grid grid-cols-[1.5fr_0.8fr_0.8fr_1fr] px-3 py-2.5 border-b border-white/[0.06] items-center">
      <span className="text-white text-[11px]">{j.name}</span>
      <span className="text-text-secondary text-[10px]">{j.type}</span>
      <span className={statusBadgeClass(j.status)}>{j.status}</span>
      <span className="text-hover text-[10px] font-semibold cursor-pointer">{j.status === 'failed' ? 'Retry' : '—'}</span>
    </div>
  ))}
</div>
```

### 8.6 Analytics

```tsx
<h1 className="text-white text-[15px] font-bold mb-3.5">Analytics</h1>
<div className="bg-card rounded-2xl p-4 mb-3.5">
  <div className="flex justify-between items-baseline mb-3"><span className="text-white text-xs font-semibold">Storage growth</span><span className="text-text-secondary text-[10px]">last 6 months</span></div>
  <div className="flex items-end gap-2 h-[70px]">
    {monthlyData.map(m => <div key={m.month} className="flex-1 bg-gradient-to-t from-primary to-accent rounded-t" style={{height: `${m.percent}%`}} />)}
  </div>
</div>
<div className="grid grid-cols-2 gap-3.5">
  <div className="bg-card rounded-2xl p-4">
    <p className="text-white text-xs font-semibold mb-3">Download trend</p>
    <div className="flex items-end gap-1.5 h-14">{trendData.map((d,i) => <div key={i} className="flex-1 bg-accent rounded-t" style={{height: `${d}%`}} />)}</div>
  </div>
  <div className="bg-card rounded-2xl p-4">
    <p className="text-white text-xs font-semibold mb-3">Most played</p>
    {mostPlayed.map(m => <div key={m.id} className="flex justify-between mb-2"><span className="text-white text-[10px]">{m.title}</span><span className="text-text-secondary text-[10px]">{m.plays}</span></div>)}
  </div>
</div>
```

### 8.7 Manage Users

```tsx
<div className="flex justify-between items-center mb-3.5">
  <h1 className="text-white text-[15px] font-bold">Users</h1>
  <button className="bg-primary text-white px-4 py-2 rounded-[10px] text-[11px] font-semibold">+ Add user</button>
</div>
<div className="bg-card rounded-2xl overflow-hidden">
  <div className="grid grid-cols-[1.5fr_1.5fr_0.8fr_0.8fr_0.8fr] px-3 py-2.5 border-b border-white/[0.06]">
    {['Nama','Email/Username','Role','Bergabung','Action'].map(h => <span key={h} className="text-text-secondary text-[9px] font-semibold uppercase">{h}</span>)}
  </div>
  {users.map(u => (
    <div key={u.id} className="grid grid-cols-[1.5fr_1.5fr_0.8fr_0.8fr_0.8fr] px-3 py-2.5 border-b border-white/[0.06] items-center">
      <span className="text-white text-[11px]">{u.name}</span>
      <span className="text-text-secondary text-[11px]">{u.username}</span>
      <span className={u.role === 'owner' ? "bg-accent/15 text-accent text-[9px] font-semibold px-2 py-0.5 rounded-lg w-fit" : "bg-white/5 text-text-secondary text-[9px] px-2 py-0.5 rounded-lg w-fit"}>{u.role}</span>
      <span className="text-text-secondary text-[10px]">{u.joinedAt}</span>
      <TrashIcon className="w-3.5 h-3.5 text-[#F87171] cursor-pointer" />
    </div>
  ))}
</div>
```
"+ Add user" buka Modal (2.6): input Username, Nama, pilihan Set Manual/Generate Password (tombol generate + tampilkan sekali dengan copy button), Role tetap "Member" (disabled, tidak bisa dipilih Owner).

### 8.8 Admin Settings

```tsx
<h1 className="text-white text-[15px] font-bold mb-4">Settings</h1>
<div className="bg-card rounded-2xl p-4 mb-3.5">
  <p className="text-white text-xs font-semibold mb-3">General</p>
  <div className="mb-3"><label className="text-text-secondary text-[10px] mb-1 block">App name</label><input value={appName} className="bg-white/5 border border-white/[0.06] rounded-lg px-3 py-2 text-white text-xs w-full max-w-[240px]" /></div>
</div>
<div className="bg-card rounded-2xl p-4">
  <p className="text-white text-xs font-semibold mb-3">Auth</p>
  <div className="flex items-center justify-between py-2 border-b border-white/[0.06]">
    <div><p className="text-white text-xs font-medium">Enable Google OAuth Login</p><p className="text-text-secondary text-[10px] mt-0.5">Tampilkan tombol Google di halaman login</p></div>
    <ToggleSwitch checked={googleOAuthEnabled} onChange={togglePatch('google_oauth_enabled')} />
  </div>
  <div className="flex items-center justify-between py-2">
    <div><p className="text-white text-xs font-medium">Maintenance Mode</p><p className="text-text-secondary text-[10px] mt-0.5">User app tampilkan halaman maintenance</p></div>
    <ToggleSwitch checked={maintenanceMode} onChange={togglePatch('maintenance_mode')} />
  </div>
</div>
```
`ToggleSwitch` component: `<button className={checked ? "w-8 h-[18px] bg-primary rounded-full relative" : "w-8 h-[18px] bg-white/15 rounded-full relative"}><span className={checked ? "absolute top-0.5 right-0.5 w-3.5 h-3.5 bg-white rounded-full" : "absolute top-0.5 left-0.5 w-3.5 h-3.5 bg-white rounded-full"} /></button>` — toggle langsung PATCH ke `/admin/settings`, tanpa perlu tombol save terpisah.

---

## 9. CATATAN AKHIR

Total 25 halaman sekarang SEMUA punya literal spec (Bagian 4, 6, 7, 8).
Tidak ada lagi halaman yang boleh diimplementasi tanpa rujukan literal di
dokumen ini. Kalau nanti ada halaman baru di luar 25 ini, TULIS dulu
spec-nya di sini (minta ke user/Claude arsitek) sebelum coding, jangan
improvisasi.

---

## 10. UPDATE TYPOGRAPHY — ganti font, perbaiki kesan "kaku"

Font lama (Inter) diganti **Plus Jakarta Sans** — alternatif gratis/legal
dengan karakter geometris-ramah mirip Spotify Circular (yang berlisensi,
tidak bisa kita pakai).

```tsx
// apps/frontend/src/app/layout.tsx dan apps/admin/src/app/layout.tsx
import { Plus_Jakarta_Sans } from "next/font/google";

const jakarta = Plus_Jakarta_Sans({
  subsets: ["latin"],
  variable: "--font-jakarta",
  weight: ["400", "500", "600", "700"],
  display: "swap",
});
```
Update `tailwind.config.ts`: `fontFamily.sans: ["var(--font-jakarta)", "sans-serif"]`

**Perbaiki juga distribusi weight** — ini penyebab kesan "kaku" yang
sebenarnya, bukan cuma soal font:
- JANGAN pakai `font-bold` (700) di banyak tempat sekaligus. Batasi HANYA
  untuk: judul halaman (Header h1), judul lagu di Now Playing/Song Detail,
  angka besar di Dashboard admin.
- Body text, label, nama lagu di list/card → `font-medium` (500) atau
  `font-semibold` (600), BUKAN `font-bold`.
- Tambahkan `tracking-tight` (letter-spacing sedikit rapat) khusus di
  heading besar (`text-lg`, `text-xl`) — ini yang bikin judul terasa
  "dirancang", bukan default browser.

## 11. DEFAULT ARTWORK & AVATAR — fallback deterministik, bukan generik

Sekarang semua placeholder artwork pakai gradient SAMA (accent→primary)
di semua tempat — makanya kesan "belum jadi/asal". Ganti jadi
**deterministik per item** (setiap lagu/album/artist dapat warna beda
tapi KONSISTEN, bukan acak tiap reload) + icon kecil di tengah.

```tsx
// lib/utils/defaultArtwork.ts
const GRADIENT_PALETTE = [
  "from-accent to-primary",
  "from-hover to-primary",
  "from-secondary to-background",
  "from-primary to-card",
  "from-hover to-secondary",
  "from-accent to-card",
];

export function getDeterministicGradient(seed: string): string {
  let hash = 0;
  for (let i = 0; i < seed.length; i++) {
    hash = seed.charCodeAt(i) + ((hash << 5) - hash);
  }
  return GRADIENT_PALETTE[Math.abs(hash) % GRADIENT_PALETTE.length];
}
```

Pemakaian (seed = id item, supaya konsisten selamanya untuk item yang sama):

```tsx
// Song/Album artwork placeholder
<div className={`w-full aspect-square rounded-xl bg-gradient-to-br ${getDeterministicGradient(song.id)} flex items-center justify-center`}>
  {!song.artworkUrl && <Music2Icon className="w-6 h-6 text-white/40" />}
</div>

// Album placeholder — icon beda
{!album.coverUrl && <DiscIcon className="w-6 h-6 text-white/40" />}

// Artist — TETAP bulat (rounded-full), bukan icon, karena manusia
<div className={`rounded-full bg-gradient-to-br ${getDeterministicGradient(artist.id)} flex items-center justify-center text-white font-bold`}>
  {!artist.avatarUrl && artist.name.charAt(0).toUpperCase()}
</div>

// User avatar — sama pola dengan artist (inisial + gradient deterministik)
<div className={`rounded-full bg-gradient-to-br ${getDeterministicGradient(user.id)} flex items-center justify-center text-white font-bold`}>
  {!user.avatarUrl && user.name.charAt(0).toUpperCase()}
</div>
```

Terapkan fallback ini ke SEMUA tempat artwork/avatar muncul — Home,
Search, Library, Playlist, Song Detail, Now Playing, Manage Songs (admin),
Manage Users (admin), Mini Player, Queue. Jangan ada satupun tempat yang
masih pakai gradient generik sama semua.

## 12. DOWNLOADS — rebuild penuh, ini sebelumnya belum benar-benar berfungsi

Masalah yang harus diperbaiki (bukan cuma tampilan, fungsinya juga):

1. **Cek dulu apakah tombol Download di Song Detail/Now Playing benar-
   benar memicu proses** — download harus: fetch audio file (via
   `/songs/:id/stream` dengan token), simpan ke **Cache API browser**
   (`caches.open('sonora-offline-v1')`), progress tracking real (pakai
   `ReadableStream` dari response body, hitung persen dari
   `Content-Length` vs bytes yang sudah diterima) — BUKAN progress dummy/
   fake timer.
2. **Cek quota storage device SEBELUM download** — panggil
   `navigator.storage.estimate()`, kalau `usage + fileSize > quota`,
   tolak download dengan pesan jelas "Storage penuh, hapus beberapa lagu
   dulu" — jangan biarkan gagal diam-diam.
3. **State download harus persisten** — simpan status per lagu (downloading/
   completed/failed) di IndexedDB (bukan cuma React state yang hilang
   kalau refresh), supaya kalau user tutup app di tengah download, waktu
   dibuka lagi statusnya tetap akurat.
4. **Tombol hapus per lagu terdownload** — harus benar-benar hapus dari
   Cache API (`cache.delete()`) DAN update IndexedDB, bukan cuma hilang
   dari UI doang.
5. **Retry untuk yang failed** — re-trigger proses nomor 1 dari awal.

Kalau ternyata SELAMA INI fitur download di frontend cuma UI kosong tanpa
logic sama sekali (kemungkinan besar, sesuai dugaan kamu "seperti belum
selesai"), BANGUN DARI NOL sesuai 5 poin di atas — jangan cuma poles
tampilan doang.

## 13. ADMIN — Lyrics Editor (baru, dari Manage Songs)

Di halaman Manage Songs (8.5... rujuk detail halaman itu), tambahkan
1 icon action baru di kolom Action tiap baris: `MicIcon` ("Lihat/Edit
Lyrics"), di samping icon edit dan delete yang sudah ada.

Klik → buka halaman baru `apps/admin/src/app/songs/[id]/lyrics/page.tsx`:

```tsx
<h1 className="text-white text-[15px] font-bold mb-1">Edit Lyrics — {songTitle}</h1>
<p className="text-text-secondary text-[11px] mb-4">{artistName} · {duration}</p>

{/* Audio player kecil untuk preview sambil edit */}
<div className="bg-card rounded-xl p-3 flex items-center gap-2.5 mb-4">
  <button onClick={togglePlay}><PlayIcon className="w-4 h-4 text-white" /></button>
  <div className="flex-1 h-1 bg-white/10 rounded-full"><div className="h-full bg-accent rounded-full" style={{width: `${progress}%`}} /></div>
  <span className="text-text-secondary text-[10px]">{currentTime}</span>
</div>

{/* Editor per baris — timestamp + text, bisa reorder/tambah/hapus baris */}
<div className="bg-card rounded-2xl p-4">
  {lines.map((line, i) => (
    <div key={line.id} className="flex items-center gap-2 mb-2">
      <input
        value={line.timestamp}
        onChange={e => updateTimestamp(line.id, e.target.value)}
        placeholder="00:12.5"
        className="w-20 bg-white/5 border border-white/[0.06] rounded-lg px-2 py-1.5 text-white text-[11px] font-mono flex-shrink-0"
      />
      <button onClick={() => setLineTimestampFromCurrentAudioTime(line.id)} title="Set dari posisi audio sekarang">
        <ClockIcon className="w-3.5 h-3.5 text-hover flex-shrink-0" />
      </button>
      <input
        value={line.text}
        onChange={e => updateText(line.id, e.target.value)}
        className="flex-1 bg-white/5 border border-white/[0.06] rounded-lg px-3 py-1.5 text-white text-xs"
      />
      <button onClick={() => deleteLine(line.id)}><XIcon className="w-3.5 h-3.5 text-[#F87171] flex-shrink-0" /></button>
    </div>
  ))}
  <button onClick={addLine} className="text-hover text-[11px] font-semibold mt-2">+ Tambah baris</button>
</div>

<button onClick={handleSave} className="bg-primary text-white rounded-2xl px-6 py-3 font-semibold text-sm mt-4">
  Simpan perubahan
</button>
```

Fitur kunci:
- **Tombol jam (⏱) di tiap baris** — klik saat audio sedang preview
  jalan, otomatis isi timestamp baris itu dengan posisi audio SEKARANG
  (`currentTime` dari audio element). Ini cara paling praktis buat sinkron
  manual tanpa hitung manual.
- Format timestamp: `MM:SS.ms` (contoh `01:23.5`)
- Simpan → `PATCH /admin/songs/:id/lyrics` dengan body array `{timestamp,
  text}[]`, backend generate ulang `synced_lrc` (format LRC standar) dari
  array itu, update kolom `lyrics.synced_lrc`.

Endpoint baru:
- `GET /admin/songs/:id/lyrics` — return lyrics existing (parse dari
  synced_lrc jadi array {timestamp, text} untuk ditampilkan di editor)
- `PATCH /admin/songs/:id/lyrics` — simpan perubahan

## 14. ADMIN — Categories: pola checklist untuk list besar vs kecil

**Genres & Moods** (~20-30 item, muat 1 layar dengan scroll wajar) —
pakai pola list biasa dengan toggle switch per baris (sama seperti
Bagian 8.3 pattern sebelumnya):

```tsx
{genres.map(g => (
  <div key={g.id} className="flex items-center justify-between py-2 border-b border-white/[0.06]">
    <span className="text-white text-xs">{g.name}</span>
    <ToggleSwitch checked={g.isActive} onChange={() => toggleGenre(g.id)} />
  </div>
))}
<button className="text-hover text-[11px] font-semibold mt-2">+ Tambah genre baru</button>
```

**Countries** (~195 item, TERLALU BANYAK untuk toggle satu-satu) — WAJIB
pakai pola search + checklist massal:

```tsx
<div className="mb-3">
  <SearchBar placeholder="Cari negara..." value={search} onChange={setSearch} />
</div>
<div className="flex items-center justify-between mb-3">
  <span className="text-text-secondary text-[11px]">{activeCount} dari {totalCount} negara aktif</span>
  <div className="flex gap-2">
    <button onClick={selectAll} className="text-hover text-[11px] font-semibold">Pilih semua</button>
    <button onClick={selectNone} className="text-hover text-[11px] font-semibold">Kosongkan</button>
  </div>
</div>
<div className="max-h-[400px] overflow-y-auto">
  {filteredCountries.map(c => (
    <label key={c.code} className="flex items-center gap-2.5 py-1.5 cursor-pointer">
      <input type="checkbox" checked={c.isActive} onChange={() => toggleCountry(c.code)}
        className="w-4 h-4 rounded accent-primary" />
      <span className="text-white text-xs">{c.name}</span>
      <span className="text-text-secondary text-[10px]">({c.code})</span>
    </label>
  ))}
</div>
<button className="bg-primary text-white rounded-xl px-5 py-2.5 text-xs font-semibold mt-3">
  Simpan perubahan
</button>
```

Checklist country TIDAK langsung PATCH per klik (beda dari genre/mood
toggle yang langsung save) — kumpulkan perubahan dulu di state lokal,
baru kirim SEKALIGUS lewat tombol "Simpan perubahan" (`PATCH
/admin/countries/bulk` dengan body `{ active_codes: string[] }`) — supaya
tidak kirim 195 request terpisah kalau admin centang banyak sekaligus.

## 15. Seed Countries — via Go seed script, bukan SQL manual

Buat `apps/backend/cmd/seed-countries/main.go`, pakai package Go yang
sudah punya data ISO 3166-1 lengkap (contoh: `github.com/biter777/countries`),
loop semua negara, insert ke tabel `countries` (code, name), default
`is_active = true` untuk SEMUA (admin yang nanti nonaktifkan yang tidak
relevan, bukan mulai dari kosong).

Dijalankan sekali manual: `go run ./cmd/seed-countries`

## 16. Migrasi data lama: albums.genre (text) → albums.genre_id (FK)

Untuk 2 lagu sample yang sudah ada (Incoherent, U TERU S...), cocokkan
`albums.genre` (text lama) ke `genres.name` yang paling mendekati
(case-insensitive match). Kalau tidak ada yang cocok persis, biarkan
`genre_id` NULL, JANGAN buat genre baru otomatis dari data lama (supaya
master data tetap terkurasi, konsisten dengan aturan di Bagian 3
sebelumnya). Tulis 1 script migrasi data sekali-jalan untuk ini, bukan
logic permanen di aplikasi.

## 17. FIX — Lyrics Editor: tombol Generate Otomatis

Di halaman Lyrics Editor (Bagian 13), tambahkan tombol baru di atas area
editor, sebelum list baris kosong:

```tsx
<div className="flex items-center justify-between mb-3">
  <p className="text-white text-xs font-semibold">Lyrics lines</p>
  <button onClick={handleAutoGenerate} disabled={isGenerating}
    className="bg-accent/10 text-accent border border-accent/30 rounded-xl px-3 py-1.5 text-[11px] font-semibold flex items-center gap-1.5">
    <SparklesIcon className="w-3.5 h-3.5" />
    {isGenerating ? "Mencari..." : "Generate Otomatis"}
  </button>
</div>
```

Endpoint baru: `POST /admin/songs/:id/lyrics/fetch` — re-trigger pencarian
lyrics ke LRCLIB pakai `songs.title` + `artists.name` TERKINI (bukan data
lama saat ingest pertama kali — penting kalau title/artist sudah dikoreksi
manual sejak itu). Response: `{ found: boolean, lines: [...] }`.

Kalau `found: false`, tampilkan pesan jelas: "Tidak ditemukan di LRCLIB
untuk '{title}' oleh '{artist}'. Coba perbaiki judul/artist dulu di
halaman edit lagu, atau tambahkan baris manual di bawah."

## 18. FIX — Duplikat icon "+" di tombol Add Song

Bug di tombol "+ Add Song" — kemungkinan besar ada 2 sumber "+" sekaligus
(icon PlusIcon DAN karakter "+" literal di teks). Cek kode tombolnya,
harus PERSIS salah satu saja:

```tsx
// BENAR — cuma 1 sumber "+"
<button className="bg-primary text-white px-4 py-2 rounded-[10px] text-[11px] font-semibold flex items-center gap-1.5">
  <PlusIcon className="w-3.5 h-3.5" /> Add Song
</button>
```
Bukan `<PlusIcon /> + Add Song` (icon DAN teks "+" sekaligus — itu yang
bikin dobel).

## 19. REBUILD — Add Song flow: analyze dulu, baru review dengan fallback manual

Alur SEKARANG (upload → langsung tersimpan apa adanya) diganti jadi
3 tahap:

**Tahap 1 — Upload file**
Modal sederhana, cuma file picker + tombol "Analisa".

**Tahap 2 — Review hasil analisa (WAJIB, jangan skip)**
Setelah backend selesai extract ID3 tag + cari genre/mood, tampilkan form
review SEBELUM benar-benar disimpan ke katalog:

```tsx
<div className="space-y-3">
  {/* Tiap field: kalau terdeteksi, tampilkan value + badge "Terdeteksi".
      Kalau TIDAK terdeteksi, tampilkan input KOSONG + border warna
      warning + label "Tidak terdeteksi, isi manual" */}

  <FieldReview
    label="Judul"
    value={detected.title}
    isDetected={!!detected.title}
    onChange={setTitleManual}
  />
  <FieldReview
    label="Artist"
    value={detected.artist}
    isDetected={!!detected.artist}
    onChange={setArtistManual}
  />
  <FieldReview label="Genre" isDetected={!!detected.genreId} type="select" options={genres} />
  <FieldReview label="Tahun rilis" isDetected={!!detected.year} type="number" />
  <FieldReview label="Negara" isDetected={!!detected.countryCode} type="select" options={countries} />
  {/* Artwork preview kalau ada embedded, atau upload manual kalau tidak */}
</div>

<button onClick={handleConfirmSave} className="bg-primary text-white rounded-2xl px-6 py-3 font-semibold text-sm w-full mt-4">
  Simpan ke katalog
</button>
```

`FieldReview` component: kalau `isDetected=false`, tambahkan
`border-[#FACC15]/40` (kuning warning) di input-nya + teks kecil di bawah
"⚠ Tidak terdeteksi dari file, isi manual" — supaya admin SADAR field mana
yang perlu dicek, bukan diam-diam kosong.

**Tahap 3 — Simpan**
Baru setelah admin klik "Simpan ke katalog", data (termasuk koreksi
manual) benar-benar masuk ke `songs` table DAN baru di titik ini trigger
proses lyrics fetch (Bagian 20) — supaya title/artist yang dipakai untuk
cari lyrics adalah yang SUDAH DIKOREKSI, bukan hasil ekstraksi mentah
yang mungkin salah.

## 20. FIX — Auto-lyrics tidak jalan saat upload baru

Investigasi WAJIB sebelum benerin:
1. Cek ingest_job_logs untuk upload terbaru — apakah step "fetch_lyrics"
   BENAR-BENAR dijalankan (ada log entry-nya) atau ter-skip total?
2. Kalau step-nya jalan tapi hasilnya kosong — cek query yang dikirim ke
   LRCLIB, apakah pakai title/artist bersih atau nama file mentah? (Ini
   kemungkinan besar akar masalahnya berdasarkan kasus ST12 di
   screenshot — title-nya "Unknown Artist" berarti query ke LRCLIB
   kemungkinan juga kacau)
3. Laporkan temuan pasti ke saya sebelum lanjut fix

Perbaikan yang PALING PENTING: pastikan step fetch_lyrics dipanggil
SETELAH Tahap 3 di Bagian 19 (setelah admin konfirmasi/koreksi title-
artist), BUKAN langsung setelah ID3 extraction mentah. Kalau title masih
"Unknown"/kosong di titik fetch lyrics dipanggil, SKIP pencarian dan
langsung set status "lyrics tidak ditemukan" (jangan buang waktu search
dengan query yang pasti gagal) — biarkan admin trigger manual lewat
tombol "Generate Otomatis" (Bagian 17) setelah title diperbaiki.
