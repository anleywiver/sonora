"use client";

import { ChevronDown } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { apiFetch } from "@/lib/api";
import { activeLineIndex, parseLRC, type LyricLine } from "@/lib/lrc";
import { cn } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface LyricsResponse {
  synced_content: string;
  plain_content: string;
}

export default function LyricsPage() {
  const router = useRouter();
  const currentSong = usePlayerStore((s) => s.currentSong);
  const positionMs = usePlayerStore((s) => s.positionMs);
  const seek = usePlayerStore((s) => s.seek);

  const [lines, setLines] = useState<LyricLine[] | null>(null);
  const [plainText, setPlainText] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);
  const lineRefs = useRef<(HTMLParagraphElement | null)[]>([]);

  useEffect(() => {
    if (!currentSong) return;
    apiFetch<LyricsResponse>(`/songs/${currentSong.id}/lyrics`)
      .then((res) => {
        if (res.synced_content) {
          setLines(parseLRC(res.synced_content));
        } else {
          setPlainText(res.plain_content || null);
        }
      })
      .catch(() => setError("Lirik tidak tersedia untuk lagu ini."));
  }, [currentSong]);

  const active = lines ? activeLineIndex(lines, positionMs) : -1;

  useEffect(() => {
    if (active >= 0) {
      lineRefs.current[active]?.scrollIntoView({ behavior: "smooth", block: "center" });
    }
  }, [active]);

  if (!currentSong) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">Tidak ada lagu yang diputar.</p>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen flex-col px-6 pb-10 pt-6">
      <div className="flex items-center justify-between">
        <button onClick={() => router.back()} aria-label="Back">
          <ChevronDown size={28} />
        </button>
        <div className="text-center">
          <p className="text-sm font-medium">{currentSong.title}</p>
          <p className="text-xs text-text-secondary">{currentSong.artistName}</p>
        </div>
        <div className="w-7" aria-hidden />
      </div>

      {error && <p className="mt-10 text-center text-sm text-error">{error}</p>}

      {lines && (
        <div className="mt-8 flex-1 space-y-4 overflow-y-auto text-center">
          {lines.map((line, i) => (
            <p
              key={i}
              ref={(el) => {
                lineRefs.current[i] = el;
              }}
              onClick={() => seek(line.timeMs)}
              className={cn(
                "cursor-pointer rounded-xl px-3 py-1 transition-opacity",
                i === active
                  ? "bg-accent/15 text-lg font-bold opacity-100"
                  : "text-base opacity-40",
              )}
              style={{ opacity: i === active ? 1 : Math.max(0.3, 1 - Math.abs(i - active) * 0.15) }}
            >
              {line.text || "♪"}
            </p>
          ))}
        </div>
      )}

      {!lines && plainText && (
        <p className="mt-8 whitespace-pre-line text-center text-base leading-relaxed">
          {plainText}
        </p>
      )}

      <p className="mt-6 text-center text-xs text-text-secondary">
        Tap baris untuk lompat ke bagian itu
      </p>
    </main>
  );
}
