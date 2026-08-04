import { create } from "zustand";

// Same in-memory-only pattern as apps/frontend/src/store/auth.ts — the
// httpOnly refresh cookie (scoped to /api/v1/auth, set by the OAuth
// callback) is what survives a reload. `role` is populated after
// bootstrap via GET /auth/me, and gates the whole admin shell: the API
// already enforces Owner-only on every /admin/* route (defense in depth),
// but the UI checks it too so a Member sees a clear "Access Denied"
// instead of a wall of failed requests.
interface AuthState {
  accessToken: string | null;
  role: string | null;
  setSession: (accessToken: string | null, role: string | null) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  accessToken: null,
  role: null,
  setSession: (accessToken, role) => set({ accessToken, role }),
}));
