"use client";

import { useEffect, useState } from "react";
import { Bar, BarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from "recharts";

import { apiFetch, ApiError } from "@/lib/api";
import { formatBytes } from "@/lib/utils";

interface TopPlayedSong {
  song_id: string;
  title: string;
  artist_name: string;
  play_count: number;
}

interface StorageGrowthPoint {
  month: string;
  total_bytes: number;
}

// Storage Growth + Most Played per docs/screens-spec.md #21. "Download
// Trend" from the same spec has no backing endpoint yet (not part of
// Sprint 11's scope, see ADR 0005) — left out rather than faked.
export default function AnalyticsPage() {
  const [topPlayed, setTopPlayed] = useState<TopPlayedSong[] | null>(null);
  const [storageGrowth, setStorageGrowth] = useState<StorageGrowthPoint[] | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      apiFetch<TopPlayedSong[]>("/admin/analytics/top-played"),
      apiFetch<StorageGrowthPoint[]>("/admin/analytics/storage-growth"),
    ])
      .then(([top, growth]) => {
        setTopPlayed(top);
        setStorageGrowth(growth);
      })
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load analytics"));
  }, []);

  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Analytics</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Storage growth (6 bulan terakhir) dan lagu paling sering diputar.
      </p>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      <div className="mt-8 rounded-card border border-border bg-card p-5">
        <h2 className="mb-4 text-sm font-semibold text-text-secondary">Storage Growth</h2>
        {!storageGrowth ? (
          <p className="text-sm text-text-secondary">Loading…</p>
        ) : (
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={storageGrowth}>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,.06)" />
                <XAxis dataKey="month" stroke="#94A3B8" fontSize={12} />
                <YAxis
                  stroke="#94A3B8"
                  fontSize={12}
                  tickFormatter={(v) => formatBytes(v)}
                  width={70}
                />
                <Tooltip
                  formatter={(value: number) => formatBytes(value)}
                  contentStyle={{ background: "#0F172A", border: "1px solid rgba(255,255,255,.06)" }}
                />
                <Bar dataKey="total_bytes" fill="#3B82F6" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        )}
      </div>

      <div className="mt-6 rounded-card border border-border bg-card p-5">
        <h2 className="mb-4 text-sm font-semibold text-text-secondary">Most Played</h2>
        {!topPlayed ? (
          <p className="text-sm text-text-secondary">Loading…</p>
        ) : topPlayed.length === 0 ? (
          <p className="text-sm text-text-secondary">Belum ada data play history.</p>
        ) : (
          <ol className="space-y-2">
            {topPlayed.map((song, i) => (
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
