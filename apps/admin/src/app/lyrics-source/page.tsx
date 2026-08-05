"use client";

import { ArrowDown, ArrowUp } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

interface Provider {
  id: string;
  name: string;
  base_url: string;
  is_enabled: boolean;
  priority: number;
  health_status: "online" | "rate_limited";
  total_lookups: number;
  match_rate_pct: number;
}

// Screens-spec #19 asks for a drag-handle reorder — with exactly one
// provider (lrclib) registered today, a drag interaction would be a
// no-op UI for a single row. Up/down buttons do the same job (swap
// priority) without pulling in a drag-and-drop library for a feature
// that has nothing to reorder yet.
export default function LyricsSourcePage() {
  const [providers, setProviders] = useState<Provider[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set());

  function load() {
    setLoading(true);
    apiFetch<Provider[]>("/admin/lyrics-providers")
      .then((list) => setProviders([...list].sort((a, b) => a.priority - b.priority)))
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load providers"))
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  function withBusy(id: string, fn: () => Promise<void>) {
    setBusyIds((s) => new Set(s).add(id));
    fn().finally(() =>
      setBusyIds((s) => {
        const next = new Set(s);
        next.delete(id);
        return next;
      }),
    );
  }

  function swapPriority(index: number, direction: -1 | 1) {
    const otherIndex = index + direction;
    if (otherIndex < 0 || otherIndex >= providers.length) return;
    const a = providers[index];
    const b = providers[otherIndex];
    withBusy(a.id, async () => {
      await Promise.all([
        apiFetch(`/admin/lyrics-providers/${a.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ priority: b.priority }),
        }),
        apiFetch(`/admin/lyrics-providers/${b.id}`, {
          method: "PATCH",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ priority: a.priority }),
        }),
      ]);
      load();
    });
  }

  function toggleEnabled(provider: Provider) {
    withBusy(provider.id, () =>
      apiFetch(`/admin/lyrics-providers/${provider.id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ is_enabled: !provider.is_enabled }),
      })
        .then(load)
        .catch((e) => setError(e instanceof ApiError ? e.message : "Update failed")),
    );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Lyrics Source</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Provider lirik, urutan prioritas, dan health status.
      </p>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      <div className="mt-8 overflow-x-auto rounded-card border border-border bg-card">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-text-secondary">
              <th className="px-4 py-3 font-medium">Priority</th>
              <th className="px-4 py-3 font-medium">Provider</th>
              <th className="px-4 py-3 font-medium">Health</th>
              <th className="px-4 py-3 font-medium">Match Rate</th>
              <th className="px-4 py-3 font-medium">Enabled</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={5} className="px-4 py-6 text-center text-text-secondary">
                  Loading…
                </td>
              </tr>
            ) : (
              providers.map((p, i) => {
                const busy = busyIds.has(p.id);
                return (
                  <tr key={p.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => swapPriority(i, -1)}
                          disabled={busy || i === 0}
                          aria-label="Move up"
                          className="text-text-secondary disabled:opacity-30"
                        >
                          <ArrowUp size={14} />
                        </button>
                        <button
                          onClick={() => swapPriority(i, 1)}
                          disabled={busy || i === providers.length - 1}
                          aria-label="Move down"
                          className="text-text-secondary disabled:opacity-30"
                        >
                          <ArrowDown size={14} />
                        </button>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <p className="font-medium capitalize">{p.name}</p>
                      <p className="text-xs text-text-secondary">{p.base_url}</p>
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={cn(
                          "rounded-full px-2.5 py-1 text-[11px] font-medium",
                          p.health_status === "online" ? "bg-success/15 text-success" : "bg-warning/15 text-warning",
                        )}
                      >
                        {p.health_status === "online" ? "Online" : "Rate limited"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-text-secondary">
                      {p.match_rate_pct < 0 ? "No data yet" : `${p.match_rate_pct.toFixed(0)}% (${p.total_lookups} lookups)`}
                    </td>
                    <td className="px-4 py-3">
                      <button
                        onClick={() => toggleEnabled(p)}
                        disabled={busy}
                        className={cn(
                          "rounded-full px-2.5 py-1 text-[11px] font-medium disabled:opacity-50",
                          p.is_enabled ? "bg-success/15 text-success" : "bg-white/5 text-text-secondary",
                        )}
                      >
                        {p.is_enabled ? "Enabled" : "Disabled"}
                      </button>
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}
