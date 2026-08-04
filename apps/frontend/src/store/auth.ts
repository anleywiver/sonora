import { create } from "zustand";

// Access token and device ID live in memory only (never localStorage) —
// the httpOnly refresh cookie is what survives a reload; see
// Providers (app/providers.tsx), which calls POST /auth/refresh on app
// start and gets both back. deviceId identifies "this browser tab's
// device" for Active Device (Sprint 8): compared against
// playback_state.active_device_id and sent to POST /ws/token.
interface AuthState {
  accessToken: string | null;
  deviceId: string | null;
  setSession: (accessToken: string | null, deviceId: string | null) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  deviceId: null,
  setSession: (accessToken, deviceId) => set({ accessToken, deviceId }),
}));
