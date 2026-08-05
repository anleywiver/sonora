# Admin Panel Standards — Pola Wajib di Semua Halaman List

Setiap halaman admin yang menampilkan TABEL/LIST data (Drive Manager, Job
Queue, Manage Users, Manage Songs) WAJIB pakai pola yang SAMA PERSIS di
bawah ini. Ini standar, bukan opsional per halaman — supaya admin panel
terasa satu produk, bukan potongan-potongan beda gaya.

---

## 1. AdminListPage Toolbar (reusable component)

```tsx
// components/admin/AdminListToolbar.tsx
interface AdminListToolbarProps {
  title: string;
  subtitle?: string;
  addButtonLabel?: string;   // kosongkan kalau halaman ini tidak butuh Add
  onAddClick?: () => void;
  searchPlaceholder?: string;
  searchValue: string;
  onSearchChange: (v: string) => void;
  filters?: FilterConfig[];
}

<div>
  {/* Header row */}
  <div className="flex items-center justify-between mb-4">
    <div>
      <h1 className="text-white text-[15px] font-bold">{title}</h1>
      {subtitle && <p className="text-text-secondary text-[11px] mt-0.5">{subtitle}</p>}
    </div>
    {addButtonLabel && (
      <button onClick={onAddClick} className="bg-primary hover:bg-hover text-white px-4 py-2 rounded-[10px] text-[11px] font-semibold flex items-center gap-1.5 transition-colors">
        <PlusIcon className="w-3.5 h-3.5" /> {addButtonLabel}
      </button>
    )}
  </div>

  {/* Search + Filter row */}
  <div className="flex items-center gap-2.5 mb-4">
    <div className="flex-1 bg-white/5 border border-white/[0.06] rounded-xl px-3 py-2 flex items-center gap-2 max-w-[320px]">
      <SearchIcon className="w-3.5 h-3.5 text-text-secondary flex-shrink-0" />
      <input
        value={searchValue}
        onChange={e => onSearchChange(e.target.value)}
        placeholder={searchPlaceholder}
        className="flex-1 bg-transparent text-white text-xs placeholder:text-text-secondary focus:outline-none"
      />
    </div>
    {filters?.map(f => <FilterDropdown key={f.key} {...f} />)}
  </div>
</div>
```

Debounce search input 300ms sebelum trigger request ke backend — jangan
fetch tiap ketikan huruf.

## 2. FilterDropdown (reusable component)

```tsx
interface FilterConfig {
  key: string;
  label: string;
  options: { value: string; label: string }[];
  selected: string | null;   // null = "All"
  onChange: (value: string | null) => void;
}

<div className="relative">
  <button onClick={toggleOpen} className="bg-white/5 border border-white/[0.06] rounded-xl px-3 py-2 flex items-center gap-1.5 text-text-secondary text-[11px]">
    {label}: <span className="text-white font-medium">{selectedLabel || 'All'}</span>
    <ChevronDownIcon className="w-3 h-3" />
  </button>
  {isOpen && (
    <div className="absolute top-full mt-1.5 left-0 bg-card border border-white/[0.06] rounded-xl p-1.5 min-w-[140px] z-10">
      <button onClick={() => onChange(null)} className="w-full text-left px-2.5 py-1.5 rounded-lg text-xs text-text-secondary hover:bg-white/5">All</button>
      {options.map(o => (
        <button key={o.value} onClick={() => onChange(o.value)} className={selected === o.value ? "w-full text-left px-2.5 py-1.5 rounded-lg text-xs bg-accent/15 text-accent font-medium" : "w-full text-left px-2.5 py-1.5 rounded-lg text-xs text-text-secondary hover:bg-white/5"}>
          {o.label}
        </button>
      ))}
    </div>
  )}
</div>
```

## 3. Pagination Footer (reusable, cocok dengan cursor pagination API kita)

```tsx
<div className="flex items-center justify-between mt-4 pt-3 border-t border-white/[0.06]">
  <span className="text-text-secondary text-[10px]">
    Menampilkan {items.length} dari {totalCount} item
  </span>
  <div className="flex gap-1.5">
    <button disabled={!hasPrevious} onClick={loadPrevious} className="px-3 py-1.5 rounded-lg text-[11px] border border-white/[0.06] text-text-secondary disabled:opacity-30 disabled:cursor-not-allowed">
      ← Sebelumnya
    </button>
    <button disabled={!hasMore} onClick={loadMore} className="px-3 py-1.5 rounded-lg text-[11px] border border-white/[0.06] text-text-secondary disabled:opacity-30 disabled:cursor-not-allowed">
      Berikutnya →
    </button>
  </div>
</div>
```

## 4. Bulk Action Bar (muncul kalau ada row terpilih via checkbox)

Tambahkan kolom checkbox di paling kiri tabel (sebelum kolom data
pertama). Saat ≥1 row dicentang, tampilkan floating bar ini:

```tsx
{selectedIds.length > 0 && (
  <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-card border border-white/[0.06] rounded-2xl px-4 py-2.5 flex items-center gap-3 shadow-lg z-20">
    <span className="text-white text-xs font-medium">{selectedIds.length} dipilih</span>
    <button onClick={handleBulkDelete} className="bg-[#F87171]/15 text-[#F87171] text-[11px] font-semibold px-3 py-1.5 rounded-lg">
      Hapus
    </button>
    <button onClick={clearSelection} className="text-text-secondary text-[11px]">Batal</button>
  </div>
)}
```

## 5. Sortable Column Header

```tsx
<th onClick={() => handleSort(columnKey)} className="cursor-pointer select-none">
  <span className="flex items-center gap-1 text-text-secondary text-[9px] font-semibold uppercase">
    {label}
    {sortKey === columnKey && (sortDir === 'asc' ? <ChevronUpIcon className="w-3 h-3" /> : <ChevronDownIcon className="w-3 h-3" />)}
  </span>
</th>
```

---

## 6. Penerapan per halaman — mana yang butuh apa

| Halaman | Add | Search | Filter | Bulk Action | Sort |
|---|---|---|---|---|---|
| Drive Manager | ✅ "+ Add drive" | ✅ nama drive | Health status (Healthy/Near full/Down) | ❌ (drive tidak di-bulk delete, terlalu berisiko) | ❌ |
| Job Queue | ❌ (job dibuat sistem) | ✅ nama job | Status (pending/running/completed/failed) DAN Type (download/crawler/jamendo) | ✅ (bulk retry untuk failed) | ✅ by created_at |
| Manage Users | ✅ "+ Add user" | ✅ nama/username | Role (Owner/Member) | ✅ (bulk hapus akses) | ✅ by joined date |
| Manage Songs | ✅ "+ Add Song" | ✅ (sudah ada) judul/artist | Genre, Storage provider, Tahun rilis (range) | ✅ (bulk hapus lagu) | ✅ by title/artist/date added |
| Lyrics Source | ❌ (provider fixed) | ❌ (list pendek) | ❌ | ❌ | ❌ (urutan by drag priority) |
| Ingest Sources | ❌ (source fixed) | ❌ | ❌ | ❌ | ❌ |
| Analytics | ❌ | ❌ | Rentang tanggal (7 hari/30 hari/90 hari) untuk chart | ❌ | ❌ |

Kolom checkbox HANYA ditambahkan ke tabel yang punya Bulk Action = ✅.
