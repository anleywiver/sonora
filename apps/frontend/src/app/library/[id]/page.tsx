"use client";

import { Play, Plus, Search as SearchIcon, X } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface PlaylistSong {
  row_id: string;
  id: string;
  title: string;
  artist_name: string;
  album_title: string;
  duration_ms: number;
}

interface PlaylistDetail {
  id: string;
  name: string;
  description: string;
  songs: PlaylistSong[];
}

interface SearchResult {
  id: string;
  title: string;
  artist_name: string;
  duration_ms: number;
}

export default function PlaylistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [playlist, setPlaylist] = useState<PlaylistDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [showAdd, setShowAdd] = useState(false);
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<SearchResult[]>([]);
  const play = usePlayerStore((s) => s.play);

  const load = () => {
    apiFetch<PlaylistDetail>(`/playlists/${id}`)
      .then(setPlaylist)
      .catch(() => setError("Playlist tidak ditemukan."));
  };

  useEffect(load, [id]);

  useEffect(() => {
    if (!query.trim()) {
      setResults([]);
      return;
    }
    const handle = setTimeout(() => {
      apiFetch<SearchResult[]>(`/search?q=${encodeURIComponent(query)}`)
        .then(setResults)
        .catch(() => setResults([]));
    }, 300);
    return () => clearTimeout(handle);
  }, [query]);

  const handleAdd = async (songId: string) => {
    try {
      await apiFetch(`/playlists/${id}/songs`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ song_id: songId }),
      });
      setQuery("");
      setResults([]);
      setShowAdd(false);
      load();
    } catch {
      setError("Gagal menambah lagu.");
    }
  };

  const handleRemove = async (rowId: string) => {
    try {
      await apiFetch(`/playlists/${id}/songs/${rowId}`, { method: "DELETE" });
      load();
    } catch {
      setError("Gagal menghapus lagu.");
    }
  };

  if (error) {
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <p className="text-sm text-error">{error}</p>
      </main>
    );
  }

  if (!playlist) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Memuat...</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <div className="mx-auto h-40 w-40 rounded-2xl bg-gradient-to-br from-accent to-primary" />
      <div className="mt-4 text-center">
        <h1 className="text-xl font-bold">{playlist.name}</h1>
        {playlist.description && (
          <p className="mt-1 text-sm text-text-secondary">{playlist.description}</p>
        )}
        <p className="mt-1 text-xs text-text-secondary">{playlist.songs.length} lagu</p>
      </div>

      <button
        onClick={() => setShowAdd((v) => !v)}
        className="mx-auto mt-4 flex items-center gap-2 rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-white"
      >
        <Plus size={16} />
        Tambah lagu
      </button>

      {showAdd && (
        <div className="mt-4">
          <div className="flex items-center gap-2 rounded-2xl border border-border bg-white/5 px-4 py-3">
            <SearchIcon size={16} className="text-text-secondary" />
            <input
              autoFocus
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Cari lagu untuk ditambahkan"
              className="w-full bg-transparent text-sm outline-none placeholder:text-text-secondary"
            />
          </div>
          <ul className="mt-2 space-y-2">
            {results.map((r) => (
              <li key={r.id}>
                <button
                  onClick={() => handleAdd(r.id)}
                  className="flex w-full items-center justify-between rounded-[18px] border border-border bg-white/5 p-3 text-left"
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{r.title}</p>
                    <p className="truncate text-xs text-text-secondary">{r.artist_name}</p>
                  </div>
                  <Plus size={16} className="text-accent" />
                </button>
              </li>
            ))}
          </ul>
        </div>
      )}

      <ul className="mt-6 space-y-2">
        {playlist.songs.map((s, i) => (
          <li
            key={s.row_id}
            className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
          >
            <span className="w-4 text-xs text-text-secondary">{i + 1}</span>
            <button
              onClick={() =>
                void play({
                  id: s.id,
                  title: s.title,
                  artistName: s.artist_name,
                  albumTitle: s.album_title,
                  durationMs: s.duration_ms,
                })
              }
              className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-white/10"
              aria-label={`Play ${s.title}`}
            >
              <Play size={14} className="ml-0.5" />
            </button>
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{s.title}</p>
              <p className="truncate text-xs text-text-secondary">{s.artist_name}</p>
            </div>
            <span className="text-xs text-text-secondary">{formatDuration(s.duration_ms)}</span>
            <button
              onClick={() => handleRemove(s.row_id)}
              aria-label={`Remove ${s.title}`}
              className="text-text-secondary"
            >
              <X size={16} />
            </button>
          </li>
        ))}
      </ul>

      {playlist.songs.length === 0 && (
        <p className="mt-6 text-center text-sm text-text-secondary">
          Belum ada lagu di playlist ini.
        </p>
      )}
    </main>
  );
}
