"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

interface RefreshResponse {
  access_token: string;
  expires_in: number;
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
  const setAccessToken = useAuthStore((s) => s.setAccessToken);
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
        if (!cancelled) setAccessToken(res.access_token);
      })
      .catch((e) => {
        if (cancelled) return;
        setAccessToken(null);
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
