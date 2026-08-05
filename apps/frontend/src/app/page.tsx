"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { usePlayerStore } from "@/store/player";

interface Me {
  avatar_url: string;
}

interface ContinueListeningItem {
  song_id: string;
  title: string;
  artist_name: string;
  progress_ms: number;
  duration_ms: number;
}

interface TrendingItem {
  id: string;
  title: string;
  artist_name: string;
  duration_ms: number;
}

export default function HomePage() {
  const [continueListening, setContinueListening] = useState<ContinueListeningItem[]>([]);
  const [trending, setTrending] = useState<TrendingItem[]>([]);
  const [me, setMe] = useState<Me | null>(null);
  const play = usePlayerStore((s) => s.play);

  useEffect(() => {
    apiFetch<ContinueListeningItem[]>("/library/continue-listening")
      .then(setContinueListening)
      .catch(() => {});
    apiFetch<TrendingItem[]>("/search/trending")
      .then(setTrending)
      .catch(() => {});
    apiFetch<Me>("/auth/me")
      .then(setMe)
      .catch(() => {});
  }, []);

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">Halo 👋</h1>
        <Link href="/profile" aria-label="Profile" className="h-9 w-9 overflow-hidden rounded-full bg-gradient-to-br from-accent to-primary">
          {me?.avatar_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={me.avatar_url} alt="" className="h-full w-full object-cover" />
          ) : null}
        </Link>
      </div>

      {continueListening.length > 0 && (
        <section className="mt-6">
          <h2 className="text-sm font-semibold text-text-secondary">Continue Listening</h2>
          <div className="mt-2 flex gap-3 overflow-x-auto pb-1">
            {continueListening.map((item) => {
              const progress = Math.min(item.progress_ms / item.duration_ms, 1) * 100;
              return (
                <Link
                  key={item.song_id}
                  href={`/song/${item.song_id}`}
                  className="w-36 flex-shrink-0 rounded-[18px] border border-border bg-white/5 p-3 backdrop-blur-md"
                >
                  <div className="h-28 w-full rounded-xl bg-gradient-to-br from-accent to-primary" />
                  <p className="mt-2 truncate text-xs font-medium">{item.title}</p>
                  <p className="truncate text-[11px] text-text-secondary">{item.artist_name}</p>
                  <div className="mt-1.5 h-1 w-full overflow-hidden rounded-full bg-white/10">
                    <div className="h-full bg-accent" style={{ width: `${progress}%` }} />
                  </div>
                </Link>
              );
            })}
          </div>
        </section>
      )}

      {trending.length > 0 && (
        <section className="mt-6">
          <h2 className="text-sm font-semibold text-text-secondary">Trending</h2>
          <ul className="mt-2 space-y-2">
            {trending.slice(0, 5).map((song) => (
              <li key={song.id}>
                <button
                  onClick={() =>
                    void play({
                      id: song.id,
                      title: song.title,
                      artistName: song.artist_name,
                      durationMs: song.duration_ms,
                    })
                  }
                  className="flex w-full items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3 text-left backdrop-blur-md"
                >
                  <div className="h-10 w-10 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{song.title}</p>
                    <p className="truncate text-xs text-text-secondary">{song.artist_name}</p>
                  </div>
                </button>
              </li>
            ))}
          </ul>
        </section>
      )}

      {continueListening.length === 0 && trending.length === 0 && (
        <p className="mt-6 text-sm text-text-secondary">
          Belum ada lagu untuk ditampilkan. Mulai dengan mencari sesuatu.
        </p>
      )}

      <Link
        href="/search"
        className="mt-6 block w-fit rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white"
      >
        Cari lagu
      </Link>
    </main>
  );
}
