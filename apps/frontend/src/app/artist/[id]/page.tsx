"use client";

import { Heart, Play } from "lucide-react";
import { useParams } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface Artist {
  id: string;
  name: string;
  image_url: string;
}

interface Song {
  id: string;
  title: string;
  duration_ms: number;
}

interface Favorite {
  target_id: string;
}

// Screens-spec #12. "Monthly listeners" and "Related artists" are left out
// deliberately — neither has a real data source (no listener-count metric,
// no artist-similarity feature), and fabricating numbers would be worse
// than a spec gap. Play + Follow and Popular songs are backed by real data.
export default function ArtistDetailPage() {
  const { id } = useParams<{ id: string }>();
  const [artist, setArtist] = useState<Artist | null>(null);
  const [songs, setSongs] = useState<Song[]>([]);
  const [isFollowing, setIsFollowing] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const play = usePlayerStore((s) => s.play);

  useEffect(() => {
    apiFetch<Artist>(`/artists/${id}`)
      .then(setArtist)
      .catch(() => setError("Artist tidak ditemukan."));
    apiFetch<Song[]>(`/artists/${id}/songs`)
      .then(setSongs)
      .catch(() => {});
    apiFetch<Favorite[]>("/favorites?type=artist")
      .then((favs) => setIsFollowing(favs.some((f) => f.target_id === id)))
      .catch(() => {});
  }, [id]);

  const toggleFollow = async () => {
    const next = !isFollowing;
    setIsFollowing(next);
    try {
      await apiFetch("/favorites", {
        method: next ? "POST" : "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ type: "artist", id }),
      });
    } catch {
      setIsFollowing(!next);
    }
  };

  const handlePlaySong = (song: Song) => {
    void play({ id: song.id, title: song.title, artistName: artist?.name ?? "", durationMs: song.duration_ms });
  };

  if (error) {
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <p className="text-sm text-error">{error}</p>
      </main>
    );
  }

  if (!artist) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Memuat...</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen pb-32">
      <div className="relative h-40 bg-gradient-to-b from-primary/40 to-background">
        <div className="absolute -bottom-12 left-1/2 h-24 w-24 -translate-x-1/2 rounded-full border-4 border-background bg-gradient-to-br from-accent to-primary" />
      </div>

      <div className="mt-16 text-center">
        <h1 className="text-xl font-bold">{artist.name}</h1>
      </div>

      <div className="mt-6 flex items-center justify-center gap-4 px-6">
        <button
          onClick={() => songs[0] && handlePlaySong(songs[0])}
          disabled={songs.length === 0}
          className="flex items-center gap-2 rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white disabled:opacity-50"
        >
          <Play size={16} className="ml-0.5" />
          Play
        </button>
        <button
          onClick={toggleFollow}
          className={
            isFollowing
              ? "flex items-center gap-2 rounded-2xl border border-accent/30 bg-accent/10 px-6 py-3 text-sm font-semibold text-accent"
              : "flex items-center gap-2 rounded-2xl border border-border bg-transparent px-6 py-3 text-sm font-semibold text-text-secondary"
          }
        >
          <Heart size={16} fill={isFollowing ? "currentColor" : "none"} />
          {isFollowing ? "Following" : "Follow"}
        </button>
      </div>

      <section className="mt-8 px-4">
        <h2 className="text-sm font-semibold text-text-secondary">Popular songs</h2>
        <ul className="mt-2 space-y-2">
          {songs.map((song, i) => (
            <li key={song.id}>
              <button
                onClick={() => handlePlaySong(song)}
                className="flex w-full items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 text-left backdrop-blur-md"
              >
                <span className="w-4 text-xs text-text-secondary">{i + 1}</span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{song.title}</p>
                </div>
                <span className="text-xs text-text-secondary">{formatDuration(song.duration_ms)}</span>
              </button>
            </li>
          ))}
        </ul>
        {songs.length === 0 && (
          <p className="mt-4 text-sm text-text-secondary">Belum ada lagu dari artist ini.</p>
        )}
      </section>
    </main>
  );
}
