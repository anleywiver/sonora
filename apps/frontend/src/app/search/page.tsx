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

interface Genre {
  id: string;
  name: string;
}

const RECENT_SEARCHES_KEY = "sonora:recent-searches";
const MAX_RECENT_SEARCHES = 8;

// Client-side only (screens-spec #4) — this is the user's own real search
// history on this device, not a fabricated/server-tracked feature. There's
// no equivalent for "Trending Searches" (that would need real aggregate
// query tracking across users, which doesn't exist yet — deliberately not
// faked with made-up terms).
function loadRecentSearches(): string[] {
  if (typeof window === "undefined") return [];
  try {
    return JSON.parse(localStorage.getItem(RECENT_SEARCHES_KEY) ?? "[]");
  } catch {
    return [];
  }
}

function saveRecentSearch(term: string) {
  const trimmed = term.trim();
  if (!trimmed) return;
  const existing = loadRecentSearches().filter((t) => t.toLowerCase() !== trimmed.toLowerCase());
  const next = [trimmed, ...existing].slice(0, MAX_RECENT_SEARCHES);
  localStorage.setItem(RECENT_SEARCHES_KEY, JSON.stringify(next));
}

export default function SearchPage() {
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SongResult[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [recent, setRecent] = useState<string[]>([]);
  const [genres, setGenres] = useState<Genre[]>([]);

  useEffect(() => {
    setRecent(loadRecentSearches());
    apiFetch<Genre[]>("/genres")
      .then(setGenres)
      .catch(() => setGenres([]));
  }, []);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }
    const handle = setTimeout(() => {
      setLoading(true);
      setError(null);
      apiFetch<SongResult[]>(`/search?q=${encodeURIComponent(query)}`)
        .then((r) => {
          setResults(r);
          if (r.length > 0) {
            saveRecentSearch(query);
            setRecent(loadRecentSearches());
          }
        })
        .catch(() => setError("Gagal mencari lagu."))
        .finally(() => setLoading(false));
    }, 300);
    return () => clearTimeout(handle);
  }, [query]);

  function clearRecent() {
    localStorage.removeItem(RECENT_SEARCHES_KEY);
    setRecent([]);
  }

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

      {!query.trim() && (
        <>
          {recent.length > 0 && (
            <section className="mt-6">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold text-text-secondary">Recent Searches</h2>
                <button onClick={clearRecent} className="text-xs text-text-secondary underline underline-offset-2">
                  Clear
                </button>
              </div>
              <div className="mt-2 flex flex-wrap gap-2">
                {recent.map((term) => (
                  <button
                    key={term}
                    onClick={() => setQuery(term)}
                    className="rounded-full border border-border px-3 py-1.5 text-xs text-text-secondary"
                  >
                    {term}
                  </button>
                ))}
              </div>
            </section>
          )}

          {genres.length > 0 && (
            <section className="mt-6">
              <h2 className="text-sm font-semibold text-text-secondary">Browse Genre</h2>
              <div className="mt-2 grid grid-cols-2 gap-3">
                {genres.map((g) => (
                  <button
                    key={g.id}
                    onClick={() => setQuery(g.name)}
                    className="rounded-[18px] bg-gradient-to-br from-accent to-primary p-4 text-left"
                  >
                    <span className="text-sm font-semibold text-white">{g.name}</span>
                  </button>
                ))}
              </div>
            </section>
          )}
        </>
      )}

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
