"use client";

import { ChevronDown, MicOff, Play } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { EmptyState } from "@/components/empty-state";
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
    <main className="flex min-h-screen flex-col bg-gradient-to-b from-card to-background px-6 pb-6 pt-6">
      <div className="flex items-center justify-between">
        <button onClick={() => router.back()} aria-label="Back">
          <ChevronDown size={22} />
        </button>
        <div className="flex items-center gap-1.5">
          <div className="h-[22px] w-[22px] flex-shrink-0 rounded-md bg-gradient-to-br from-hover to-primary" />
          <span className="text-[10px] font-semibold text-white/70">{currentSong.title}</span>
        </div>
        <div className="w-[18px]" aria-hidden />
      </div>

      {error && (
        <EmptyState icon={MicOff} message="Lirik tidak tersedia untuk lagu ini." />
      )}

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

      {(lines || plainText) && (
        <div className="mt-3.5 flex items-center justify-between border-t border-white/[0.08] pt-3.5">
          <span className="text-[10px] text-white/50">Tap baris untuk lompat ke bagian itu</span>
          <div className="flex items-center gap-1.5">
            <Play size={14} className="text-white" />
            <span className="text-[10px] text-white/50">Auto-scroll</span>
          </div>
        </div>
      )}
    </main>
  );
}
