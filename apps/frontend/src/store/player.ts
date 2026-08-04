import { create } from "zustand";

import { apiFetch } from "@/lib/api";

export interface PlayerSong {
  id: string;
  title: string;
  artistName: string;
  albumTitle?: string;
  albumCoverUrl?: string;
  durationMs: number;
}

interface StreamTokenResponse {
  token: string;
  expires_in: number;
}

interface PlayerState {
  currentSong: PlayerSong | null;
  isPlaying: boolean;
  isLoading: boolean;
  positionMs: number;
  durationMs: number;
  error: string | null;
  play: (song: PlayerSong) => Promise<void>;
  togglePlay: () => void;
  seek: (ms: number) => void;
}

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

// A single <audio> element lives on window so navigating between pages
// (Search -> Song Detail -> Now Playing) never interrupts playback.
function getAudio(): HTMLAudioElement {
  const w = window as unknown as { __sonoraAudio?: HTMLAudioElement };
  if (!w.__sonoraAudio) {
    w.__sonoraAudio = new Audio();
  }
  return w.__sonoraAudio;
}

// Fire-and-forget: a failed history write shouldn't interrupt playback.
function recordHistory(songId: string, progressMs: number) {
  apiFetch("/history", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ song_id: songId, progress_ms: Math.round(progressMs) }),
  }).catch(() => {});
}

export const usePlayerStore = create<PlayerState>((set) => ({
  currentSong: null,
  isPlaying: false,
  isLoading: false,
  positionMs: 0,
  durationMs: 0,
  error: null,

  play: async (song) => {
    set({ isLoading: true, error: null, currentSong: song });
    try {
      const { token } = await apiFetch<StreamTokenResponse>(`/songs/${song.id}/stream-token`, {
        method: "POST",
      });

      const audio = getAudio();
      audio.src = `${API_BASE}/songs/${song.id}/stream?token=${token}`;
      audio.ontimeupdate = () => set({ positionMs: audio.currentTime * 1000 });
      audio.onloadedmetadata = () => set({ durationMs: audio.duration * 1000 });
      audio.onplay = () => set({ isPlaying: true });
      // Recorded on pause/end rather than continuously — a discrete event
      // per listening session is enough for Continue Listening to work,
      // without spamming POST /history on every timeupdate tick.
      audio.onpause = () => {
        set({ isPlaying: false });
        recordHistory(song.id, audio.currentTime * 1000);
      };
      audio.onended = () => recordHistory(song.id, audio.currentTime * 1000);
      audio.onerror = () =>
        set({ error: "Gagal memutar lagu ini.", isPlaying: false, isLoading: false });

      await audio.play();
      set({ isLoading: false });
    } catch (e) {
      // Never surface the raw browser/DOMException message here — it leaks
      // technical wording ("Failed to load because no supported source was
      // found") straight into user-facing UI. Always show the same
      // friendly message; the real cause goes to the console instead.
      console.error("player: play failed", e);
      set({
        isLoading: false,
        isPlaying: false,
        error: "Gagal memutar lagu ini.",
      });
    }
  },

  togglePlay: () => {
    const audio = getAudio();
    if (audio.paused) {
      void audio.play();
    } else {
      audio.pause();
    }
  },

  seek: (ms) => {
    const audio = getAudio();
    audio.currentTime = ms / 1000;
    set({ positionMs: ms });
  },
}));
