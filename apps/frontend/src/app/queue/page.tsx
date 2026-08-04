"use client";

import { ChevronDown, X } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface QueueItem {
  id: string;
  song_id: string;
  title: string;
  artist_name: string;
  duration_ms: number;
}

export default function QueuePage() {
  const router = useRouter();
  const currentSong = usePlayerStore((s) => s.currentSong);
  const [items, setItems] = useState<QueueItem[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    apiFetch<QueueItem[]>("/queue")
      .then(setItems)
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleRemove = async (id: string) => {
    await apiFetch(`/queue/${id}`, { method: "DELETE" });
    load();
  };

  const handleClear = async () => {
    await apiFetch("/queue", { method: "DELETE" });
    load();
  };

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <div className="flex items-center justify-between">
        <button onClick={() => router.back()} aria-label="Back">
          <ChevronDown size={28} />
        </button>
        <h1 className="text-lg font-bold">Queue</h1>
        <button onClick={handleClear} className="text-xs text-text-secondary">
          Clear queue
        </button>
      </div>

      {currentSong && (
        <div className="mt-6">
          <p className="text-xs text-text-secondary">Now Playing</p>
          <div className="mt-2 flex items-center gap-3 rounded-[18px] border border-border bg-accent/10 p-3.5">
            <div className="h-10 w-10 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
            <div className="min-w-0">
              <p className="truncate text-sm font-medium">{currentSong.title}</p>
              <p className="truncate text-xs text-text-secondary">{currentSong.artistName}</p>
            </div>
          </div>
        </div>
      )}

      <p className="mt-6 text-xs text-text-secondary">Next up</p>
      {loading && <p className="mt-2 text-sm text-text-secondary">Memuat...</p>}
      <ul className="mt-2 space-y-2">
        {items.map((item) => (
          <li
            key={item.id}
            className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5"
          >
            <div className="h-10 w-10 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{item.title}</p>
              <p className="truncate text-xs text-text-secondary">{item.artist_name}</p>
            </div>
            <span className="text-xs text-text-secondary">{formatDuration(item.duration_ms)}</span>
            <button onClick={() => handleRemove(item.id)} aria-label={`Remove ${item.title}`}>
              <X size={16} className="text-text-secondary" />
            </button>
          </li>
        ))}
      </ul>

      {!loading && items.length === 0 && (
        <p className="mt-4 text-sm text-text-secondary">Queue kosong.</p>
      )}
    </main>
  );
}
