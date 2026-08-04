import { create } from "zustand";

import { apiFetch } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

const WS_BASE = (process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1")
  .replace(/^http/, "ws")
  .replace(/\/api\/v1$/, "");

interface WSTokenResponse {
  token: string;
}

interface ServerMessage {
  type: "player:state" | "player:command";
  data: unknown;
}

interface PlayerStateData {
  active_device_id: string | null;
  current_song_id: string | null;
  position_ms: number;
  is_playing: boolean;
}

interface WSState {
  connected: boolean;
  activeDeviceId: string | null;
  connect: () => Promise<void>;
}

let socket: WebSocket | null = null;

export const useWSStore = create<WSState>((set) => ({
  connected: false,
  activeDeviceId: null,

  connect: async () => {
    const { deviceId } = useAuthStore.getState();
    if (!deviceId || socket) return;

    try {
      const { token } = await apiFetch<WSTokenResponse>("/ws/token", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ device_id: deviceId }),
      });

      const ws = new WebSocket(`${WS_BASE}/ws?token=${token}`);
      socket = ws;

      ws.onopen = () => set({ connected: true });
      ws.onclose = () => {
        socket = null;
        set({ connected: false });
      };
      ws.onerror = () => {
        socket = null;
        set({ connected: false });
      };
      ws.onmessage = (event) => {
        let msg: ServerMessage;
        try {
          msg = JSON.parse(event.data);
        } catch {
          return;
        }
        if (msg.type === "player:state") {
          const data = msg.data as PlayerStateData;
          set({ activeDeviceId: data.active_device_id });
        }
        // player:command is only actionable by the Active Device with a
        // real <audio> element to control — see docs/decisions/0003 for
        // why full remote-control command execution isn't wired into the
        // player UI yet (untestable without real multi-device audio;
        // backend relay itself is verified independently).
      };
    } catch {
      // No connectivity right now — the app still works without
      // real-time sync, just without cross-device updates.
    }
  },
}));

// Claims this device as Active and syncs state — called by the player
// store after a local play/pause/seek so other devices see it happen.
export async function syncRemoteState(
  currentSongId: string,
  positionMs: number,
  isPlaying: boolean,
) {
  const { deviceId } = useAuthStore.getState();
  if (!deviceId) return;
  try {
    await apiFetch("/player/state", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        active_device_id: deviceId,
        current_song_id: currentSongId,
        position_ms: Math.round(positionMs),
        is_playing: isPlaying,
      }),
    });
  } catch {
    // Best-effort — local playback (usePlayerStore) is unaffected.
  }
}

export function isThisDeviceActive(): boolean {
  const { deviceId } = useAuthStore.getState();
  const { activeDeviceId } = useWSStore.getState();
  return !!deviceId && deviceId === activeDeviceId;
}
