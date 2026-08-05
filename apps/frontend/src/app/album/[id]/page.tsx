"use client";

import { Play, Shuffle, Download, Heart } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface AlbumSong {
  id: string;
  title: string;
  duration_ms: number;
  track_number: number | null;
}

interface AlbumDetail {
  id: string;
  title: string;
  cover_url: string;
  artist_id: string;
  artist_name: string;
  released_at: string | null;
  songs: AlbumSong[];
}

interface Favorite {
  target_id: string;
}

// Screens-spec #13. "Genre" is left out — the schema only associates
// genres with individual songs (song_genres), not albums, so there's no
// real per-album genre to show without fabricating one.
export default function AlbumDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [album, setAlbum] = useState<AlbumDetail | null>(null);
  const [favoriteSongIds, setFavoriteSongIds] = useState<Set<string>>(new Set());
  const [error, setError] = useState<string | null>(null);
  const play = usePlayerStore((s) => s.play);

  useEffect(() => {
    apiFetch<AlbumDetail>(`/albums/${id}`)
      .then(setAlbum)
      .catch(() => setError("Album tidak ditemukan."));
    apiFetch<Favorite[]>("/favorites?type=song")
      .then((favs) => setFavoriteSongIds(new Set(favs.map((f) => f.target_id))))
      .catch(() => {});
  }, [id]);

  const playSong = (song: AlbumSong) => {
    if (!album) return;
    void play({
      id: song.id,
      title: song.title,
      artistName: album.artist_name,
      albumTitle: album.title,
      albumCoverUrl: album.cover_url,
      durationMs: song.duration_ms,
    });
  };

  const playAlbum = (shuffle: boolean) => {
    if (!album || album.songs.length === 0) return;
    const songs = shuffle ? [...album.songs].sort(() => Math.random() - 0.5) : album.songs;
    playSong(songs[0]);
  };

  const toggleFavoriteSong = async (songId: string) => {
    const isFav = favoriteSongIds.has(songId);
    setFavoriteSongIds((prev) => {
      const next = new Set(prev);
      if (isFav) next.delete(songId);
      else next.add(songId);
      return next;
    });
    try {
      await apiFetch("/favorites", {
        method: isFav ? "DELETE" : "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type: "song", id: songId }),
      });
    } catch {
      setFavoriteSongIds((prev) => {
        const next = new Set(prev);
        if (isFav) next.add(songId);
        else next.delete(songId);
        return next;
      });
    }
  };

  if (error) {
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <p className="text-sm text-error">{error}</p>
      </main>
    );
  }

  if (!album) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Memuat...</p>
      </main>
    );
  }

  const year = album.released_at ? new Date(album.released_at).getFullYear() : null;

  return (
    <main className="relative min-h-screen overflow-hidden px-6 pb-32 pt-10">
      {/* Hero gradient (screens-spec #13) — see song/[id]/page.tsx for why
          this is a fixed tint rather than per-cover color extraction. */}
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-gradient-to-b from-primary/25 to-transparent"
        aria-hidden
      />
      <div className="relative mx-auto h-56 w-56 rounded-2xl bg-gradient-to-br from-accent to-primary" aria-hidden />
      <div className="mt-6 text-center">
        <h1 className="text-xl font-bold">{album.title}</h1>
        <Link
          href={`/artist/${album.artist_id}`}
          className="mt-1 inline-block text-sm text-text-secondary hover:text-text-primary"
        >
          {album.artist_name}
        </Link>
        <p className="mt-1 text-xs text-text-secondary">
          {[year, `${album.songs.length} lagu`].filter(Boolean).join(" · ")}
        </p>
      </div>

      <div className="mt-6 flex items-center justify-center gap-6">
        <button
          onClick={() => playAlbum(true)}
          aria-label="Shuffle play"
          className="text-text-secondary"
        >
          <Shuffle size={20} />
        </button>
        <button
          onClick={() => playAlbum(false)}
          disabled={album.songs.length === 0}
          className="flex h-14 w-14 items-center justify-center rounded-full bg-primary text-white disabled:opacity-50"
          aria-label="Play album"
        >
          <Play size={24} className="ml-1" />
        </button>
        <button aria-label="Download album" className="text-text-secondary" disabled>
          <Download size={20} />
        </button>
      </div>

      <ul className="mt-8 space-y-2">
        {album.songs.map((song, i) => {
          const isFav = favoriteSongIds.has(song.id);
          return (
            <li
              key={song.id}
              className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
            >
              <span className="w-4 text-xs text-text-secondary">{song.track_number ?? i + 1}</span>
              <button
                onClick={() => playSong(song)}
                className="min-w-0 flex-1 text-left"
                aria-label={`Play ${song.title}`}
              >
                <p className="truncate text-sm font-medium">{song.title}</p>
              </button>
              <span className="text-xs text-text-secondary">{formatDuration(song.duration_ms)}</span>
              <button
                onClick={() => toggleFavoriteSong(song.id)}
                aria-label={isFav ? "Unfavorite" : "Favorite"}
                className={isFav ? "text-error" : "text-text-secondary"}
              >
                <Heart size={16} fill={isFav ? "currentColor" : "none"} />
              </button>
            </li>
          );
        })}
      </ul>

      {album.songs.length === 0 && (
        <p className="mt-6 text-center text-sm text-text-secondary">Belum ada lagu di album ini.</p>
      )}
    </main>
  );
}
