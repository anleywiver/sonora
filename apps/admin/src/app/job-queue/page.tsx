export default function JobQueuePage() {
  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Job Queue</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Ingest job table (status, retry). Endpoint admin job listing belum dibangun.
      </p>

      <div className="mt-8 rounded-card border border-border bg-card p-5">
        <p className="text-sm text-text-secondary">Belum ada data. Fitur ini menyusul.</p>
      </div>
    </main>
  );
}
