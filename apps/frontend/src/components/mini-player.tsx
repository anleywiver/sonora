"use client";

import { Play, Pause } from "lucide-react";
import Link from "next/link";

import { usePlayerStore } from "@/store/player";

export function MiniPlayer() {
  const currentSong = usePlayerStore((s) => s.currentSong);
  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const error = usePlayerStore((s) => s.error);
  const togglePlay = usePlayerStore((s) => s.togglePlay);

  if (!currentSong) return null;

  return (
    <div className="fixed bottom-16 left-3 right-3 z-30 mx-auto flex max-w-md items-center gap-3 rounded-2xl border border-border bg-white/6 px-3 py-2 backdrop-blur-md">
      <div className="h-10 w-10 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
      <Link href="/now-playing" className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{currentSong.title}</p>
        <p className="truncate text-xs text-text-secondary">
          {error ?? currentSong.artistName}
        </p>
      </Link>
      <button
        onClick={togglePlay}
        className="flex h-9 w-9 flex-shrink-0 items-center justify-center rounded-full bg-white text-background"
        aria-label={isPlaying ? "Pause" : "Play"}
      >
        {isPlaying ? <Pause size={18} /> : <Play size={18} className="ml-0.5" />}
      </button>
    </div>
  );
}
