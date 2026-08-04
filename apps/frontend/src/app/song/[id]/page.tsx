"use client";

import { Play, Pause, Heart } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface SongDetail {
  id: string;
  title: string;
  duration_ms: number;
  artist_id: string;
  artist_name: string;
  album_id: string | null;
  album_title: string;
  album_cover_url: string;
}

interface Favorite {
  target_id: string;
}

export default function SongDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [song, setSong] = useState<SongDetail | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [isFavorite, setIsFavorite] = useState(false);

  const currentSong = usePlayerStore((s) => s.currentSong);
  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const isLoading = usePlayerStore((s) => s.isLoading);
  const playerError = usePlayerStore((s) => s.error);
  const play = usePlayerStore((s) => s.play);
  const togglePlay = usePlayerStore((s) => s.togglePlay);

  useEffect(() => {
    apiFetch<SongDetail>(`/songs/${id}`)
      .then(setSong)
      .catch(() => setError("Lagu tidak ditemukan."));
    apiFetch<Favorite[]>("/favorites?type=song")
      .then((favs) => setIsFavorite(favs.some((f) => f.target_id === id)))
      .catch(() => {});
  }, [id]);

  const toggleFavorite = async () => {
    const next = !isFavorite;
    setIsFavorite(next); // optimistic
    try {
      if (next) {
        await apiFetch("/favorites", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ type: "song", id }),
        });
      } else {
        await apiFetch("/favorites", {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ type: "song", id }),
        });
      }
    } catch {
      setIsFavorite(!next); // revert on failure
    }
  };

  if (error) {
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <p className="text-sm text-error">{error}</p>
      </main>
    );
  }

  if (!song) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Memuat...</p>
      </main>
    );
  }

  const isCurrent = currentSong?.id === song.id;

  const handlePlay = () => {
    if (isCurrent) {
      togglePlay();
    } else {
      void play({
        id: song.id,
        title: song.title,
        artistName: song.artist_name,
        albumTitle: song.album_title,
        durationMs: song.duration_ms,
      });
    }
  };

  return (
    <main className="min-h-screen px-6 pb-32 pt-10">
      <div
        className="mx-auto h-56 w-56 rounded-2xl bg-gradient-to-br from-accent to-primary"
        aria-hidden
      />
      <div className="mt-6 text-center">
        <h1 className="text-xl font-bold">{song.title}</h1>
        <p className="mt-1 text-sm text-text-secondary">{song.artist_name}</p>
        {song.album_title && (
          <p className="mt-0.5 text-xs text-text-secondary">{song.album_title}</p>
        )}
        <p className="mt-1 text-xs text-text-secondary">{formatDuration(song.duration_ms)}</p>
      </div>

      <div className="mt-8 flex items-center justify-center gap-6">
        <button
          onClick={toggleFavorite}
          aria-label={isFavorite ? "Unfavorite" : "Favorite"}
          className={isFavorite ? "text-error" : "text-text-secondary"}
        >
          <Heart size={22} fill={isFavorite ? "currentColor" : "none"} />
        </button>
        <button
          onClick={handlePlay}
          disabled={isLoading}
          className="flex h-14 w-14 items-center justify-center rounded-full bg-primary text-white disabled:opacity-50"
          aria-label={isCurrent && isPlaying ? "Pause" : "Play"}
        >
          {isCurrent && isPlaying ? <Pause size={24} /> : <Play size={24} className="ml-1" />}
        </button>
        <div className="w-[22px]" aria-hidden />
      </div>

      {isCurrent && playerError && (
        <p className="mt-4 text-center text-sm text-error">{playerError}</p>
      )}
    </main>
  );
}
