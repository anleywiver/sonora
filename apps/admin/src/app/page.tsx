"use client";

import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { formatBytes, cn } from "@/lib/utils";

interface StorageDistributionItem {
  id: string;
  label: string;
  used_bytes: number;
  quota_bytes: number | null;
  used_pct: number;
}

interface DashboardData {
  total_songs: number;
  total_users: number;
  total_drives: number;
  total_storage_bytes: number;
  storage_distribution: StorageDistributionItem[];
  background_jobs: Record<string, number>;
}

interface TopPlayedSong {
  song_id: string;
  title: string;
  artist_name: string;
  play_count: number;
}

export default function AdminDashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null);
  const [topPlayed, setTopPlayed] = useState<TopPlayedSong[]>([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<DashboardData>("/admin/dashboard")
      .then(setData)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load dashboard"));
    apiFetch<TopPlayedSong[]>("/admin/analytics/top-played")
      .then(setTopPlayed)
      .catch(() => {});
  }, []);

  const stats = [
    { label: "Total Song", value: data?.total_songs },
    { label: "Total User", value: data?.total_users },
    { label: "Storage Terpakai", value: data ? formatBytes(data.total_storage_bytes) : undefined },
    { label: "Drive Aktif", value: data?.total_drives },
  ];

  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <p className="mt-1 text-sm text-text-secondary">Ringkasan sistem Sonora.</p>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      <div className="mt-8 grid grid-cols-2 gap-4 lg:grid-cols-4">
        {stats.map(({ label, value }) => (
          <div key={label} className="rounded-card border border-border bg-card p-5">
            <p className="text-xs text-text-secondary">{label}</p>
            <p className="mt-2 text-2xl font-bold">{value ?? "—"}</p>
          </div>
        ))}
      </div>

      <div className="mt-6 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <div className="rounded-card border border-border bg-card p-5">
          <h2 className="mb-4 text-sm font-semibold text-text-secondary">Storage Distribution</h2>
          {!data ? (
            <p className="text-sm text-text-secondary">Loading…</p>
          ) : data.storage_distribution.length === 0 ? (
            <p className="text-sm text-text-secondary">Belum ada storage account.</p>
          ) : (
            <div className="space-y-3">
              {data.storage_distribution.map((d) => (
                <div key={d.id}>
                  <div className="mb-1 flex justify-between text-xs">
                    <span>{d.label}</span>
                    <span className="text-text-secondary">
                      {formatBytes(d.used_bytes)}
                      {d.quota_bytes ? ` / ${formatBytes(d.quota_bytes)}` : ""}
                    </span>
                  </div>
                  <div className="h-1.5 w-full overflow-hidden rounded-full bg-white/10">
                    <div
                      className={cn("h-full rounded-full", d.used_pct > 90 ? "bg-error" : "bg-accent")}
                      style={{ width: `${Math.min(100, d.used_pct)}%` }}
                    />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        <div className="rounded-card border border-border bg-card p-5">
          <h2 className="mb-4 text-sm font-semibold text-text-secondary">Background Jobs</h2>
          {!data ? (
            <p className="text-sm text-text-secondary">Loading…</p>
          ) : (
            <div className="grid grid-cols-2 gap-3">
              {["pending", "processing", "completed", "failed"].map((status) => (
                <div key={status} className="rounded-control border border-border p-3">
                  <p className="text-xs capitalize text-text-secondary">{status}</p>
                  <p className="mt-1 text-lg font-bold">{data.background_jobs[status] ?? 0}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      <div className="mt-6 rounded-card border border-border bg-card p-5">
        <h2 className="mb-4 text-sm font-semibold text-text-secondary">Top Played</h2>
        {topPlayed.length === 0 ? (
          <p className="text-sm text-text-secondary">Belum ada data.</p>
        ) : (
          <ol className="space-y-2">
            {topPlayed.slice(0, 5).map((song, i) => (
              <li key={song.song_id} className="flex items-center justify-between text-sm">
                <span>
                  <span className="text-text-secondary">{i + 1}.</span> {song.title}{" "}
                  <span className="text-xs text-text-secondary">— {song.artist_name}</span>
                </span>
                <span className="text-xs text-text-secondary">{song.play_count}x</span>
              </li>
            ))}
          </ol>
        )}
      </div>
    </main>
  );
}
