import { create } from "zustand";

import { apiFetch } from "@/lib/api";
import { getDownload } from "@/lib/offline-db";
import { syncRemoteState } from "@/store/ws";

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

interface QueueItem {
  id: string;
  song_id: string;
  title: string;
  artist_name: string;
  duration_ms: number;
}

export type RepeatMode = "off" | "all" | "one";

// A song is only "skippable back to" once we've actually moved on from it —
// capped so this in-memory stack (lost on reload, same as the rest of
// playback state) can't grow unbounded during a long session.
const MAX_PREVIOUS_SONGS = 20;

interface PlayerState {
  currentSong: PlayerSong | null;
  isPlaying: boolean;
  isLoading: boolean;
  positionMs: number;
  durationMs: number;
  error: string | null;
  shuffleEnabled: boolean;
  repeatMode: RepeatMode;
  playbackRate: number;
  previousSongs: PlayerSong[];
  play: (song: PlayerSong) => Promise<void>;
  togglePlay: () => void;
  seek: (ms: number) => void;
  toggleShuffle: () => void;
  cycleRepeat: () => void;
  cyclePlaybackRate: () => void;
  playNext: () => Promise<void>;
  playPrevious: () => Promise<void>;
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

export const usePlayerStore = create<PlayerState>((set, get) => ({
  currentSong: null,
  isPlaying: false,
  isLoading: false,
  positionMs: 0,
  durationMs: 0,
  error: null,
  shuffleEnabled: false,
  repeatMode: "off",
  playbackRate: 1,
  previousSongs: [],

  play: async (song) => {
    const prevSong = get().currentSong;
    set((state) => ({
      isLoading: true,
      error: null,
      currentSong: song,
      previousSongs:
        prevSong && prevSong.id !== song.id
          ? [prevSong, ...state.previousSongs].slice(0, MAX_PREVIOUS_SONGS)
          : state.previousSongs,
    }));
    try {
      const audio = getAudio();

      // Play from the offline download if we have one — no network round
      // trip at all, and it works with no connectivity.
      const downloaded = await getDownload(song.id);
      if (downloaded) {
        audio.src = URL.createObjectURL(downloaded.blob);
      } else {
        const { token } = await apiFetch<StreamTokenResponse>(`/songs/${song.id}/stream-token`, {
          method: "POST",
        });
        audio.src = `${API_BASE}/songs/${song.id}/stream?token=${token}`;
      }
      audio.playbackRate = get().playbackRate;
      audio.ontimeupdate = () => set({ positionMs: audio.currentTime * 1000 });
      audio.onloadedmetadata = () => set({ durationMs: audio.duration * 1000 });
      audio.onplay = () => set({ isPlaying: true });
      // Recorded on pause/end rather than continuously — a discrete event
      // per listening session is enough for Continue Listening to work,
      // without spamming POST /history on every timeupdate tick.
      audio.onpause = () => {
        set({ isPlaying: false });
        recordHistory(song.id, audio.currentTime * 1000);
        void syncRemoteState(song.id, audio.currentTime * 1000, false);
      };
      audio.onended = () => {
        recordHistory(song.id, audio.currentTime * 1000);
        // Repeat semantics (screens-spec / ui-implementation-spec #6.2):
        // "one" restarts the same track. Both "off" and "all" advance
        // through the queue the same way — the queue here is a flat,
        // manually-managed list (not tied to an album/playlist context),
        // so there's no defined "start" to loop back to once it empties;
        // that's a deliberate simplification, not an oversight.
        if (get().repeatMode === "one") {
          audio.currentTime = 0;
          void audio.play();
        } else {
          void get().playNext();
        }
      };
      audio.onerror = () =>
        set({ error: "Gagal memutar lagu ini.", isPlaying: false, isLoading: false });

      await audio.play();
      set({ isLoading: false });
      // Pressing play "here" claims this device as Active (Sprint 8) — the
      // explicit device switcher (Now Playing → Devices) is for moving
      // playback to a *different* device without touching this one.
      void syncRemoteState(song.id, audio.currentTime * 1000, true);
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
    const song = get().currentSong;
    if (song) {
      void syncRemoteState(song.id, ms, !audio.paused);
    }
  },

  toggleShuffle: () => set((s) => ({ shuffleEnabled: !s.shuffleEnabled })),

  cycleRepeat: () =>
    set((s) => ({
      repeatMode: s.repeatMode === "off" ? "all" : s.repeatMode === "all" ? "one" : "off",
    })),

  cyclePlaybackRate: () => {
    const rates = [1, 1.25, 1.5, 2];
    const current = get().playbackRate;
    const next = rates[(rates.indexOf(current) + 1) % rates.length];
    getAudio().playbackRate = next;
    set({ playbackRate: next });
  },

  // Pulls from the real /queue (application/queue, Sprint 6) — the item
  // played is removed from the queue same as if the user had tapped it
  // directly, not a separate "auto-advance" concept.
  playNext: async () => {
    try {
      const queue = await apiFetch<QueueItem[]>("/queue");
      if (queue.length === 0) return;
      const index = get().shuffleEnabled ? Math.floor(Math.random() * queue.length) : 0;
      const item = queue[index];
      await get().play({
        id: item.song_id,
        title: item.title,
        artistName: item.artist_name,
        durationMs: item.duration_ms,
      });
      await apiFetch(`/queue/${item.id}`, { method: "DELETE" }).catch(() => {});
    } catch {
      // No queue / request failed — same end state as reaching the end of
      // a track with nothing queued: playback just stops.
    }
  },

  // Spotify-style behavior: restart the current track if more than a few
  // seconds in, only actually go back a track otherwise.
  playPrevious: async () => {
    const { positionMs, previousSongs } = get();
    if (positionMs > 3000 || previousSongs.length === 0) {
      get().seek(0);
      return;
    }
    const [prev, ...rest] = previousSongs;
    set({ previousSongs: rest });
    await get().play(prev);
  },
}));
