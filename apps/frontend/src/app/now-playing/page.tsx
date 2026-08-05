"use client";

import {
  ChevronDown,
  ListMusic,
  Mic2,
  Moon,
  Pause,
  Play,
  Repeat,
  Repeat1,
  Shuffle,
  SkipBack,
  SkipForward,
  Speaker,
} from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { cn, formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

const SLEEP_TIMER_OPTIONS_MIN = [15, 30, 45, 60];

export default function NowPlayingPage() {
  const router = useRouter();
  const currentSong = usePlayerStore((s) => s.currentSong);
  const isPlaying = usePlayerStore((s) => s.isPlaying);
  const positionMs = usePlayerStore((s) => s.positionMs);
  const durationMs = usePlayerStore((s) => s.durationMs);
  const error = usePlayerStore((s) => s.error);
  const togglePlay = usePlayerStore((s) => s.togglePlay);
  const seek = usePlayerStore((s) => s.seek);
  const shuffleEnabled = usePlayerStore((s) => s.shuffleEnabled);
  const repeatMode = usePlayerStore((s) => s.repeatMode);
  const playbackRate = usePlayerStore((s) => s.playbackRate);
  const toggleShuffle = usePlayerStore((s) => s.toggleShuffle);
  const cycleRepeat = usePlayerStore((s) => s.cycleRepeat);
  const cyclePlaybackRate = usePlayerStore((s) => s.cyclePlaybackRate);
  const playNext = usePlayerStore((s) => s.playNext);
  const playPrevious = usePlayerStore((s) => s.playPrevious);

  const [sleepMinutes, setSleepMinutes] = useState<number | null>(null);
  const [showSleepMenu, setShowSleepMenu] = useState(false);
  const sleepTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Sleep Timer is a page-local, client-only concept (ui-implementation-
  // spec #6.2/#4.5) — not persisted, not synced to other devices, same
  // as how Speed is a per-session control.
  function setSleepTimer(minutes: number | null) {
    if (sleepTimeoutRef.current) clearTimeout(sleepTimeoutRef.current);
    setSleepMinutes(minutes);
    setShowSleepMenu(false);
    if (minutes !== null) {
      sleepTimeoutRef.current = setTimeout(
        () => {
          if (usePlayerStore.getState().isPlaying) togglePlay();
          setSleepMinutes(null);
        },
        minutes * 60 * 1000,
      );
    }
  }

  useEffect(() => {
    return () => {
      if (sleepTimeoutRef.current) clearTimeout(sleepTimeoutRef.current);
    };
  }, []);

  if (!currentSong) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Tidak ada lagu yang diputar.</p>
      </main>
    );
  }

  const totalMs = durationMs || currentSong.durationMs;
  const clampedPosition = Math.min(positionMs, totalMs || positionMs);

  const RepeatIcon = repeatMode === "one" ? Repeat1 : Repeat;

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

      <div className="mt-6 flex items-center justify-center gap-5">
        <button
          onClick={toggleShuffle}
          aria-label="Shuffle"
          aria-pressed={shuffleEnabled}
          className={shuffleEnabled ? "text-hover" : "text-text-secondary"}
        >
          <Shuffle size={18} />
        </button>
        <button onClick={() => void playPrevious()} aria-label="Previous">
          <SkipBack size={22} fill="currentColor" />
        </button>
        <button
          onClick={togglePlay}
          className="flex h-[52px] w-[52px] items-center justify-center rounded-full bg-white text-background"
          aria-label={isPlaying ? "Pause" : "Play"}
        >
          {isPlaying ? <Pause size={26} /> : <Play size={26} className="ml-1" />}
        </button>
        <button onClick={() => void playNext()} aria-label="Next">
          <SkipForward size={22} fill="currentColor" />
        </button>
        <button
          onClick={cycleRepeat}
          aria-label="Repeat"
          aria-pressed={repeatMode !== "off"}
          className={repeatMode !== "off" ? "text-hover" : "text-text-secondary"}
        >
          <RepeatIcon size={18} />
        </button>
      </div>

      <div className="relative mt-8 grid grid-cols-5 gap-2 border-t border-white/10 pt-3.5 text-text-secondary">
        <div className="relative flex flex-col items-center gap-1 text-[10px]">
          <button onClick={() => setShowSleepMenu((v) => !v)} aria-label="Sleep Timer" className="flex flex-col items-center gap-1">
            <Moon size={18} className={sleepMinutes ? "text-hover" : undefined} />
            {sleepMinutes ? `${sleepMinutes}m` : "Sleep"}
          </button>
          {showSleepMenu && (
            <div className="absolute bottom-full left-1/2 mb-2 w-32 -translate-x-1/2 rounded-2xl border border-border bg-card p-2 text-left shadow-lg">
              {SLEEP_TIMER_OPTIONS_MIN.map((m) => (
                <button
                  key={m}
                  onClick={() => setSleepTimer(m)}
                  className="block w-full rounded-lg px-2 py-1.5 text-xs text-text-primary hover:bg-white/5"
                >
                  {m} menit
                </button>
              ))}
              <button
                onClick={() => setSleepTimer(null)}
                className="block w-full rounded-lg px-2 py-1.5 text-xs text-text-secondary hover:bg-white/5"
              >
                Off
              </button>
            </div>
          )}
        </div>

        <button
          onClick={cyclePlaybackRate}
          aria-label="Playback speed"
          className={cn("flex flex-col items-center gap-1 text-[10px]", playbackRate !== 1 && "text-hover")}
        >
          <span className="text-sm font-semibold leading-none">{playbackRate}x</span>
          Speed
        </button>

        <Link href="/lyrics" aria-label="Lyrics" className="flex flex-col items-center gap-1 text-[10px] text-hover">
          <Mic2 size={18} />
          Lyrics
        </Link>

        <Link href="/queue" aria-label="Queue" className="flex flex-col items-center gap-1 text-[10px]">
          <ListMusic size={18} />
          Queue
        </Link>

        <Link href="/devices" aria-label="Devices" className="flex flex-col items-center gap-1 text-[10px]">
          <Speaker size={18} />
          Devices
        </Link>
      </div>
    </main>
  );
}
