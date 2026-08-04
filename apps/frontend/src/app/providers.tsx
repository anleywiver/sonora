"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";
import { useWSStore } from "@/store/ws";

interface RefreshResponse {
  access_token: string;
  expires_in: number;
  device_id: string;
}

const PUBLIC_PATHS = ["/login", "/auth/callback"];

// Bootstraps the session on every full page load: the access token only
// ever lives in memory, so a reload has none — but the httpOnly refresh
// cookie (set by the OAuth callback) can silently mint a new one without
// sending the user back through Google.
//
// Skipped entirely on /login and /auth/callback: the root layout (and
// this provider with it) persists across client-side navigation, so a
// refresh call started here while landing on /auth/callback can resolve
// *after* that page has already set the real token from the URL fragment
// — its rejection handler would then null out a token that was just set
// correctly. Bootstrapping never applies to those two pages anyway.
export function Providers({ children }: { children: React.ReactNode }) {
  const setSession = useAuthStore((s) => s.setSession);
  const [ready, setReady] = useState(false);
  const pathname = usePathname();
  const router = useRouter();

  useEffect(() => {
    if (PUBLIC_PATHS.includes(pathname)) {
      setReady(true);
      return;
    }

    let cancelled = false;

    apiFetch<RefreshResponse>("/auth/refresh", { method: "POST" })
      .then((res) => {
        if (!cancelled) {
          setSession(res.access_token, res.device_id);
          void useWSStore.getState().connect();
        }
      })
      .catch((e) => {
        if (cancelled) return;
        setSession(null, null);
        if (e instanceof ApiError) {
          router.replace("/login");
        }
      })
      .finally(() => {
        if (!cancelled) setReady(true);
      });

    return () => {
      cancelled = true;
    };
    // Deliberately keyed to the path at first mount only, not every
    // navigation — see the comment above.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!ready) {
    return null;
  }

  return <>{children}</>;
}
