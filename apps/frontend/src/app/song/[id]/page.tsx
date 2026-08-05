"use client";

import { Play, Pause, Heart, ListPlus, Download, Check } from "lucide-react";
import Link from "next/link";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { getDownload, saveDownload } from "@/lib/offline-db";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

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
    getDownload(id).then((d) => setIsDownloaded(!!d));
  }, [id]);

  const [isDownloaded, setIsDownloaded] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [downloadError, setDownloadError] = useState<string | null>(null);
  const handleDownload = async () => {
    if (!song) return;
    setDownloading(true);
    setDownloadError(null);
    try {
      const { token } = await apiFetch<{ token: string }>(`/songs/${id}/stream-token`, {
        method: "POST",
      });
      const res = await fetch(`${API_BASE}/songs/${id}/stream?token=${token}`);
      if (!res.ok) throw new Error("download failed");
      const blob = await res.blob();
      await saveDownload(id, {
        blob,
        mimeType: blob.type,
        title: song.title,
        artistName: song.artist_name,
        sizeBytes: blob.size,
        downloadedAt: Date.now(),
      });
      setIsDownloaded(true);
    } catch {
      setDownloadError("Gagal download lagu ini.");
    } finally {
      setDownloading(false);
    }
  };

  const [queued, setQueued] = useState(false);
  const [queueError, setQueueError] = useState<string | null>(null);
  const handleAddToQueue = async () => {
    try {
      await apiFetch("/queue", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ song_id: id }),
      });
      setQueued(true);
      setTimeout(() => setQueued(false), 2000);
    } catch {
      setQueueError("Gagal menambah ke queue.");
    }
  };

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
    <main className="relative min-h-screen overflow-hidden px-6 pb-32 pt-10">
      {/* Hero gradient (screens-spec #6) — fixed blue tint rather than
          per-image color extraction, same fallback approach Now Playing
          already uses (design-system.md), since artwork today is a
          placeholder gradient anyway, not a real cover image. */}
      <div
        className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-gradient-to-b from-primary/25 to-transparent"
        aria-hidden
      />
      <div
        className="relative mx-auto h-56 w-56 rounded-2xl bg-gradient-to-br from-accent to-primary"
        aria-hidden
      />
      <div className="mt-6 text-center">
        <h1 className="text-xl font-bold">{song.title}</h1>
        <Link
          href={`/artist/${song.artist_id}`}
          className="mt-1 inline-block text-sm text-text-secondary hover:text-text-primary"
        >
          {song.artist_name}
        </Link>
        {song.album_title && (
          <p className="mt-0.5">
            <Link
              href={`/album/${song.album_id}`}
              className="text-xs text-text-secondary hover:text-text-primary"
            >
              {song.album_title}
            </Link>
          </p>
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
        <button
          onClick={handleAddToQueue}
          aria-label="Add to queue"
          className={queued ? "text-success" : "text-text-secondary"}
        >
          <ListPlus size={22} />
        </button>
        <button
          onClick={handleDownload}
          disabled={downloading || isDownloaded}
          aria-label={isDownloaded ? "Downloaded" : "Download"}
          className={isDownloaded ? "text-success" : "text-text-secondary disabled:opacity-50"}
        >
          {isDownloaded ? <Check size={22} /> : <Download size={22} />}
        </button>
      </div>

      {isCurrent && playerError && (
        <p className="mt-4 text-center text-sm text-error">{playerError}</p>
      )}
      {queueError && <p className="mt-4 text-center text-sm text-error">{queueError}</p>}
      {queued && <p className="mt-4 text-center text-sm text-success">Ditambahkan ke queue.</p>}
      {downloadError && <p className="mt-4 text-center text-sm text-error">{downloadError}</p>}
      {isDownloaded && (
        <p className="mt-4 text-center text-sm text-success">Tersedia offline.</p>
      )}
    </main>
  );
}
