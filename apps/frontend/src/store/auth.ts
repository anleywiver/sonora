import { create } from "zustand";

// Access token lives in memory only (never localStorage) — the httpOnly
// refresh cookie is what survives a reload; see bootstrapSession() in
// providers.tsx, which calls POST /auth/refresh on app start.
interface AuthState {
  accessToken: string | null;
  setAccessToken: (token: string | null) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  setAccessToken: (token) => set({ accessToken: token }),
}));
