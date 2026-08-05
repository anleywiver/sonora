# Design Techniques Playbook — Sonora

Dokumen ini menjelaskan TEKNIK di balik setiap efek visual di design system
kita — bukan cuma "pakai warna X", tapi KENAPA dan BAGAIMANA supaya kamu
paham cara kerjanya, bisa terapkan dengan benar, dan bisa reproduksi pola
yang sama untuk komponen baru di masa depan.

Baca ini SEBELUM implementasi ulang halaman manapun. Kalau ada bagian yang
tidak jelas, tanya dulu — jangan tebak.

---

## 1. Bottom Navigation — kenapa melengkung seperti pil, bukan kotak nempel

Bug yang sering terjadi: nav nempel penuh di tepi layar, radius kecil atau
cuma di sudut atas, warna solid. Hasilnya jadi kotak biasa, bukan "pil
melayang" seperti di desain.

3 hal yang HARUS jalan bersamaan:

1. **Jarak dari tepi layar** — `mx-3 mb-2` (bukan `inset-x-0 bottom-0`
   langsung nempel). Margin ini yang bikin dia "melayang".
2. **Radius besar di SEMUA sudut** — `rounded-[22px]`. Karena tinggi bar
   cuma ~50px, radius 22px itu proporsinya besar → efek jadi kapsul/pil,
   bukan kotak bersudut tumpul.
3. **Background semi-transparan + blur** — `bg-card/85 backdrop-blur-md`
   (bukan warna solid). Ini yang bikin dia terlihat "kaca buram melayang
   di atas konten".

Hilangkan salah satu dari 3 ini → efek pil hilang.

## 2. Glassmorphism Card — kenapa terlihat "kaca buram"

BUKAN cuma `background: putih transparan`. Butuh 3 layer bersamaan:

```css
background: rgba(255,255,255, 0.05-0.06);  /* sangat tipis */
border: 1px solid rgba(255,255,255, 0.06-0.08);
backdrop-filter: blur(10px);  /* PALING PENTING — blur konten DI BELAKANG elemen */
-webkit-backdrop-filter: blur(10px);  /* wajib untuk Safari */
```

Tanpa `backdrop-filter`, ini cuma kotak putih transparan biasa — efek
"kaca" hilang total. Dipakai di: card lagu/album, mini player, bottom nav,
modal/bottom sheet.

## 3. Artwork Placeholder — gradasi diagonal, bukan warna solid

```css
background: linear-gradient(135deg, #3B82F6, #1D4ED8);
```

Sudut 135° = diagonal kiri-atas ke kanan-bawah. Bikin artwork terlihat
"hidup" walau belum ada gambar asli. Kombinasi warna SELALU dari keluarga
biru kita (accent→primary, hover→primary, dst) — jangan pernah keluar
palet.

## 4. Hero Gradient (Song Detail, Playlist, Album, Artist) — vertikal, berhenti di 45%

```css
background: linear-gradient(to bottom, #1D4ED8 0%, #050816 45%);
```

Gradient VERTIKAL (beda dari artwork yang diagonal). Berhenti di 45%
(bukan 100%) → transisi ke gelap terjadi cepat, menciptakan efek "sorotan
panggung" di area artwork/judul, baru mereda jadi gelap di tracklist.
Kalau stop di 100%, transisi terlalu lambat dan terlihat aneh.

## 5. Radial Glow (Splash, Now Playing background) — BEDA dari hero gradient

```css
background: radial-gradient(circle at 50% 20%, #1D4ED8 0%, #050816 65%);
```

Ini `radial-gradient` (pancaran dari 1 titik ke segala arah), BUKAN
`linear-gradient`. `at 50% 20%` = titik pusat cahaya di tengah horizontal,
20% dari atas — supaya sorotan jatuh ke arah artwork.

**PENTING**: Tailwind TIDAK PUNYA utility class untuk radial gradient.
WAJIB pakai inline style:
```tsx
style={{ background: 'radial-gradient(circle at 50% 20%, #1D4ED8 0%, #050816 65%)' }}
```
Jangan coba paksa pakai className Tailwind untuk ini.

## 6. Avatar "Menggigit" Banner (Artist Detail) — teknik overlap

```css
.banner { position: relative; }  /* wadah HARUS relative */
.avatar {
  position: absolute;
  bottom: -30px;  /* NEGATIF — ini yang bikin avatar "keluar" dari banner */
  border: 3px solid #050816;  /* warna = BACKGROUND HALAMAN, bukan warna lain */
}
```

Border avatar harus SEWARNA background halaman — jadi terlihat seperti
"cincin pemisah" antara avatar dan banner, seolah avatar dipotong keluar.

## 7. Animasi Splash — glow pulse + dots bounce

```css
@keyframes glow {
  0%, 100% { opacity: 0.55; transform: scale(1); }
  50% { opacity: 0.85; transform: scale(1.08); }
}
```

Animasi membesar-mengecil DAN opacity berubah BERSAMAAN → efek "bernafas",
bukan cuma kedip. Durasi 2.4 detik (lambat) = kesan tenang/premium.

Dots loading: 3 titik, delay animasi beda 0.2 detik per titik → efek
"menjalar" satu-satu, bukan berkedip bersamaan.

**WAJIB** dibungkus:
```css
@media (prefers-reduced-motion: no-preference) {
  /* animasi di sini */
}
```
Supaya user yang matikan animasi di setting OS tidak dipaksa lihat — ini
aksesibilitas, bukan kosmetik opsional.

## 8. Drag Handle Queue — affordance, bukan dekorasi

Icon `grip-vertical` (titik-titik vertikal) + `cursor: grab` saat hover.
Icon ini SECARA UNIVERSAL dikenali user sebagai "bisa di-drag" — jangan
ganti dengan icon lain (misal titik 3 horizontal/menu), user tidak akan
tahu itu draggable. Functional drag pakai `@dnd-kit/sortable`, icon ini
cuma visual cue.

## 9. Tab Selector — "pil di dalam pil"

```
Container luar: rounded-[14px], padding 4px, bg putih 4% transparan
Tab aktif: rounded-[10px] DI DALAM container, bg biru solid
```

Radius tab aktif (10px) HARUS lebih kecil dari container (14px) — supaya
ada "napas" 4px di semua sisi, seperti pil yang pas masuk slot. Radius
sama besar → sudut "bentrok", terlihat aneh.

## 10. Tombol Play Besar — icon gelap di atas tombol putih

Tombol lingkaran putih solid, icon play/pause di dalamnya warna
**background halaman** (`#050816`), BUKAN putih/biru. Alasan: kontras —
icon putih di atas tombol putih tidak kelihatan sama sekali.

## 11. Lyrics — highlight baris aktif "meluber" ke luar teks

```css
.active-line {
  background: rgba(59,130,246, 0.15);
  border-radius: 12px;
  padding: 4px 10px;
  margin-left: -10px;  /* NEGATIF, sama besar dengan padding kiri */
}
```

`margin-left: -10px` menarik elemen ke kiri sejumlah padding yang
ditambahkan → highlight rata dengan teks sekitar (tidak menjorok ke
kanan), walau sebenarnya punya padding. Trick yang sering kelewat.

## 12. Section Header — 2 gaya, jangan tertukar

- **Section biasa** ("Recently played", "Trending now"): `text-xs
  font-semibold text-white` — normal case.
- **Section dengan micro-label** ("TOP RESULT", "IN PROGRESS"): `text-[10px]
  font-semibold uppercase tracking-wide text-text-secondary` — huruf
  besar semua, warna abu-abu (BUKAN putih). Kesan "label kategori
  teknis", beda dari judul section biasa.

## 13. Progress Bar / Seekbar — bukan `<input type="range">` polos

```css
Track: height 3-5px, bg rgba(255,255,255,0.10-0.15), rounded-full
Fill: height 100% dari track, bg warna aktif, rounded-full, width = %
```

Track DAN fill sama-sama `rounded-full` — supaya ujung fill yang
"terpotong" di tengah tetap terlihat rounded, bukan kotak.

## 14. Divider Antar Baris List — garis super tipis

```css
border-bottom: 1px solid rgba(255,255,255, 0.06);
```

Warna PERSIS sama dengan token `--border`. Terapkan ke SEMUA item KECUALI
item terakhir dalam grup (cek index terakhir, atau CSS `:not(:last-child)`).

## 15. Horizontal Scroll TANPA scrollbar kelihatan

```css
overflow-x: auto;
scrollbar-width: none; /* Firefox */
-ms-overflow-style: none; /* IE/Edge */
&::-webkit-scrollbar { display: none; } /* Chrome/Safari */
```

Semua row card (Recently Played, Trending) scrollable tapi scrollbar
disembunyikan — kesan app native. SERING KELUPAAN, cek manual setelah
implementasi.

## 16. Avatar Bulat vs Artwork Kotak — aturan tetap

- **`rounded-full`**: SELALU untuk foto orang (Avatar user, foto Artist).
  Manusia = organik = bulat.
- **`rounded-xl`/`rounded-2xl`**: SELALU untuk artwork lagu/album/playlist.
  Produk/objek = kotak rounded.

Di Search Result "Top Result": kalau hasilnya Artist → bulat, kalau
hasilnya Song → kotak rounded. Style IKUT jenis data, bukan template fix.

## 17. Card "Active/Highlighted" — tint + border lebih pekat

```css
background: rgba(59,130,246, 0.10);   /* accent, opacity rendah */
border: 1px solid rgba(59,130,246, 0.25);   /* accent, opacity lebih tinggi */
```

Border SELALU lebih pekat dari background (0.25 vs 0.10) — outline jelas
tapi tetap "menyatu", bukan kotak solid berwarna terpisah.

## 18. Chip/Pill untuk Tag — beda dari Badge status

Badge = untuk STATUS (healthy/failed/pending), lihat Bagian 1.4 di
`ui-implementation-spec.md`. Chip = untuk TAG/FILTER yang bisa diklik:

```css
background: rgba(59,130,246, 0.12);
color: #93C5FD;  /* accent muda, BUKAN accent solid */
padding: 6px 12px;
border-radius: 9999px;  /* pill penuh */
font-size: 11px;
```

Chip WAJIB clickable (cursor pointer + hover state), Badge tidak perlu.

## 19. Divider "OR" — garis mengapit teks

```tsx
<div className="flex items-center gap-2.5">
  <div className="flex-1 h-px bg-white/[0.06]" />
  <span className="text-text-secondary text-[11px]">or continue with username</span>
  <div className="flex-1 h-px bg-white/[0.06]" />
</div>
```

Flexbox dengan 2 elemen garis (`flex-1 h-px`) mengapit teks di tengah —
BUKAN border-top di teks dengan padding, hasilnya beda secara visual.

## 20. Icon Highlight dalam Grup — 1 warna beda sebagai penanda

Di utility row Now Playing (Sleep Timer, Speed, Lyrics, Queue): SEMUA icon
warna sama (putih transparan) KECUALI 1 yang paling relevan konteksnya
(Lyrics dikasih warna `#60A5FA`) — menonjol sebagai "aksi disarankan".
Ini pola SENGAJA, bukan inkonsistensi yang perlu diseragamkan.

---

## Cara pakai dokumen ini

Setiap kali implementasi komponen yang menyebutkan salah satu istilah di
atas (glassmorphism, hero gradient, radial glow, drag handle, dst), BUKA
dokumen ini, cari nomor teknik yang relevan, ikuti PERSIS caranya —
termasuk detail kecil seperti radius yang berbeda antara container dan
child, atau margin negatif. Detail kecil itu yang membedakan hasil "mirip"
dengan hasil "identik".
