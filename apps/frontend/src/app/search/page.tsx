"use client";

import { Search as SearchIcon } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";

interface SongResult {
  id: string;
  title: string;
  artist_name: string;
  album_title: string;
  duration_ms: number;
}

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SongResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }
    const handle = setTimeout(() => {
      setLoading(true);
      setError(null);
      apiFetch<SongResult[]>(`/search?q=${encodeURIComponent(query)}`)
        .then(setResults)
        .catch(() => setError("Gagal mencari lagu."))
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(handle);
  }, [query]);

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <div className="flex items-center gap-2 rounded-2xl border border-border bg-white/5 px-4 py-3">
        <SearchIcon size={18} className="text-text-secondary" />
        <input
          autoFocus
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder="Cari lagu, artis, atau album"
          className="w-full bg-transparent text-sm outline-none placeholder:text-text-secondary"
        />
      </div>

      {loading && <p className="mt-6 text-sm text-text-secondary">Mencari...</p>}
      {error && <p className="mt-6 text-sm text-error">{error}</p>}

      <ul className="mt-4 space-y-2">
        {results.map((song) => (
          <li key={song.id}>
            <Link
              href={`/song/${song.id}`}
              className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
            >
              <div className="h-12 w-12 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{song.title}</p>
                <p className="truncate text-xs text-text-secondary">{song.artist_name}</p>
              </div>
              <span className="text-xs text-text-secondary">
                {formatDuration(song.duration_ms)}
              </span>
            </Link>
          </li>
        ))}
      </ul>

      {!loading && query.trim() && results.length === 0 && (
        <p className="mt-6 text-sm text-text-secondary">
          Tidak ada hasil untuk &quot;{query}&quot;.
        </p>
      )}
    </main>
  );
}
