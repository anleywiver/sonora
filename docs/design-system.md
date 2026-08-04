# Design System — Sonora

Ini spesifikasi visual lengkap dari sesi desain (STEP 1-8). Claude Code
HARUS mengikuti ini persis saat implementasi komponen — jangan improvisasi
warna, radius, atau spacing baru.

## Color Tokens

- background: #050816
- card: #0F172A
- primary: #1D4ED8
- secondary: #2563EB
- accent: #3B82F6
- hover: #60A5FA
- text-primary: #FFFFFF
- text-secondary: #94A3B8
- border: rgba(255,255,255,.06)

Warna semantik (dipakai TERBATAS, cuma untuk status — bukan dekoratif):

- success: #4ADE80 (downloaded, healthy, completed)
- warning: #FACC15 (downloading, pending, degraded)
- error: #F87171 (failed, near full, down)
- info: #60A5FA (badge "new", info state)

## Typography

Font: **Inter**, weight 400/500/600/700, lewat `next/font/google`.

- Heading XL: 28px / 700 / letter-spacing -0.01em
- Heading M: 20px / 600
- Body strong: 16px / 500
- Body: 14px / 400
- Body secondary: 14px / 400 / color text-secondary
- Caption: 11-12px / 500-600

## Spacing Scale

`4 8 12 16 24 32 48` (px) — pakai skala Tailwind default yang cocok
(`p-1 p-2 p-3 p-4 p-6 p-8 p-12`).

## Radius

- Button/Input/control: 16px (rounded-2xl)
- Card: 20px (rounded-[20px])
- Modal/BottomSheet top: 24px (rounded-t-3xl)
- Artwork kecil: 12px
- Artwork besar: 16-20px
- Avatar/circular: 9999px (rounded-full)

## Style Rules

- **Dark mode only** — tidak ada light mode di v1.
- **Glassmorphism ringan**: `backdrop-blur-md` (bukan blur berat) +
  `bg-white/5` atau `bg-white/6`, dipakai HANYA di: card lagu/album, mini
  player, bottom navigation, modal/bottom sheet. JANGAN taruh blur di semua
  elemen — itu yang bikin "norak" seperti yang eksplisit dihindari di brief
  awal.
- **1 accent action per screen** — cuma 1 tombol primary biru solid per
  layar (biasanya Play), sisanya secondary/ghost.
- **Icon**: Lucide React (`lucide-react`), ukuran 16-24px tergantung konteks.
- **Animasi**: Framer Motion untuk page transition (fade + slight scale),
  hero animation untuk artwork (shared layout transition dari list ke
  detail), ripple untuk button press.

## Component Spec

### Button

- Primary: `bg-primary text-white rounded-2xl px-6 py-3 font-semibold`,
  hover → `bg-hover`
- Secondary: `bg-accent/10 text-accent border border-accent/30 rounded-2xl`
- Ghost: `bg-transparent text-text-secondary border border-border`

### Card (album/playlist/song)

- `bg-white/5 border border-border rounded-[18px] p-3.5 backdrop-blur-md`
- Artwork: aspect-square, `rounded-xl`, gradient placeholder kalau belum
  load: `bg-gradient-to-br from-accent to-primary`

### Badge

- `rounded-full px-3 py-1 text-[11px] font-semibold`
- Warna sesuai status semantik di atas, background versi 15% opacity dari
  warna solid-nya (misal `bg-accent/15 text-accent`)

### Bottom Navigation

- Fixed bottom, `bg-card/85 backdrop-blur-md rounded-[22px] mx-3 mb-2 px-5
  py-2.5`
- 5 item: Home, Search, Library, Favorite, Settings — icon aktif warna
  `hover` (#60A5FA), tidak aktif `text-secondary`

### Mini Player

- Fixed di atas bottom nav, `bg-white/6 border border-border rounded-2xl
  backdrop-blur-md px-3 py-2`
- Artwork 38-44px, title+artist truncate, icon Transfer Playback
  (device-tv) + Heart + Play/Pause

### Now Playing (full screen)

- Background: `radial-gradient` dari warna dominan artwork (fallback ke
  `radial-gradient(circle at 50% 20%, primary 0%, background 65%)`)
- Artwork besar center, seekbar tipis (4px), tombol kontrol besar di
  tengah (52px circle putih untuk play/pause)
- Bottom utility row: Sleep Timer, Speed, Lyrics, Queue — bukan Like/Share
  di sini (itu ada di Song Detail)

### Lyrics (fullscreen)

- Baris aktif: `text-lg font-bold bg-accent/15 rounded-xl px-3` — baris
  lain fade progresif (`opacity-30` s.d. `opacity-55` tergantung jarak dari
  baris aktif)
- Tap baris → seek audio ke timestamp itu

### Bottom Sheet / Modal

- Drag handle bar di atas (`w-9 h-1 bg-white/20 rounded-full mx-auto`)
- `bg-card border-t border-border rounded-t-3xl p-5`, overlay `bg-black/50`

### Skeleton Loading

- `bg-white/6` sampai `bg-white/8` shimmer, radius mengikuti elemen aslinya

## Referensi implementasi

Semua komponen ini pernah didesain visual sebagai mockup interaktif di sesi
chat sebelum coding dimulai — kalau user (pemilik project) share screenshot
referensi, cocokkan persis dengan itu, dokumen ini adalah rangkumannya
dalam bentuk spec tertulis.
