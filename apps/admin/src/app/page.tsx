export default function AdminDashboardPage() {
  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Ringkasan sistem — stat aggregation endpoint belum dibangun, lihat
        Drive Manager untuk data storage yang sudah live.
      </p>

      <div className="mt-8 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {["Total Song", "Total User", "Storage Terpakai", "Drive Aktif"].map((label) => (
          <div key={label} className="rounded-card border border-border bg-card p-5">
            <p className="text-xs text-text-secondary">{label}</p>
            <p className="mt-2 text-2xl font-bold">—</p>
          </div>
        ))}
      </div>
    </main>
  );
}
