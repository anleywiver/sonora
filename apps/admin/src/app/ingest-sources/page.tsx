const providers = [
  { name: "Manual Upload", status: "connected" },
  { name: "Bandcamp", status: "disconnected" },
  { name: "Cloud Sync", status: "disconnected" },
];

// List provider legal (Bandcamp, Cloud Sync) — BUKAN keyword/auto-download
// generic crawler. Lihat docs/screens-spec.md catatan koreksi "Crawler" ->
// "Ingest Sources". Provider connect flow (OAuth Bandcamp, Drive/Dropbox
// cloud sync setup) belum dibangun — placeholder read-only, functional di
// sprint ingest lanjutan.
export default function IngestSourcesPage() {
  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Ingest Sources</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Sumber ingest legal — manual upload, Bandcamp, cloud sync. Tidak ada scraping/auto-download.
      </p>

      <div className="mt-8 space-y-3">
        {providers.map((p) => (
          <div
            key={p.name}
            className="flex items-center justify-between rounded-card border border-border bg-card p-4"
          >
            <span className="font-medium">{p.name}</span>
            <span
              className={
                p.status === "connected"
                  ? "rounded-full bg-success/15 px-2.5 py-1 text-[11px] font-medium text-success"
                  : "rounded-full bg-white/5 px-2.5 py-1 text-[11px] font-medium text-text-secondary"
              }
            >
              {p.status}
            </span>
          </div>
        ))}
      </div>
    </main>
  );
}
