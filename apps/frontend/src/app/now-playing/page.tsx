"use client";

import { ChevronDown, ListMusic, Mic2, Pause, Play } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";

import { formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

export default function NowPlayingPage() {
  const router = useRouter();
  const currentSong = usePlayerStore((s) => s.currentSong);
  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const positionMs = usePlayerStore((s) => s.positionMs);
  const durationMs = usePlayerStore((s) => s.durationMs);
  const error = usePlayerStore((s) => s.error);
  const togglePlay = usePlayerStore((s) => s.togglePlay);
  const seek = usePlayerStore((s) => s.seek);

  if (!currentSong) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Tidak ada lagu yang diputar.</p>
      </main>
    );
  }

  const totalMs = durationMs || currentSong.durationMs;
  const clampedPosition = Math.min(positionMs, totalMs || positionMs);

  return (
    <main
      className="flex min-h-screen flex-col px-6 pb-10 pt-6"
      style={{
        background: "radial-gradient(circle at 50% 20%, #1D4ED8 0%, #050816 65%)",
      }}
    >
      <button onClick={() => router.back()} aria-label="Back" className="self-start">
        <ChevronDown size={28} />
      </button>

      <div className="mx-auto mt-10 h-72 w-72 rounded-2xl bg-gradient-to-br from-accent to-primary" />

      <div className="mt-8 text-center">
        <h1 className="text-xl font-bold">{currentSong.title}</h1>
        <p className="mt-1 text-sm text-text-secondary">{currentSong.artistName}</p>
      </div>

      {error && <p className="mt-4 text-center text-sm text-error">{error}</p>}

      <div className="mt-8">
        <input
          type="range"
          min={0}
          max={totalMs || 1}
          value={clampedPosition}
          onChange={(e) => seek(Number(e.target.value))}
          className="w-full accent-white"
        />
        <div className="mt-1 flex justify-between text-xs text-text-secondary">
          <span>{formatDuration(clampedPosition)}</span>
          <span>{formatDuration(totalMs)}</span>
        </div>
      </div>

      <div className="mt-6 flex justify-center">
        <button
          onClick={togglePlay}
          className="flex h-[52px] w-[52px] items-center justify-center rounded-full bg-white text-background"
          aria-label={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? <Pause size={26} /> : <Play size={26} className="ml-1" />}
        </button>
      </div>

      <div className="mt-8 flex items-center justify-center gap-10 text-text-secondary">
        <Link href="/lyrics" aria-label="Lyrics" className="flex flex-col items-center gap-1 text-xs">
          <Mic2 size={20} />
          Lyrics
        </Link>
        <Link href="/queue" aria-label="Queue" className="flex flex-col items-center gap-1 text-xs">
          <ListMusic size={20} />
          Queue
        </Link>
      </div>
    </main>
  );
}
