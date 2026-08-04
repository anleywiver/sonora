"use client";

import { useEffect, useState } from "react";
import { usePathname, useRouter } from "next/navigation";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

interface RefreshResponse {
  access_token: string;
}

interface MeResponse {
  role: string;
}

const PUBLIC_PATHS = ["/login", "/auth/callback"];

// Same bootstrap pattern as apps/frontend/src/app/providers.tsx — skipped
// on /login and /auth/callback for the same reason (a refresh started here
// could resolve after the callback page already set the real token from
// the URL fragment and clobber it). Additionally fetches /auth/me for the
// role, since the admin shell gates on Owner.
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
      .then(async (res) => {
        if (cancelled) return;
        setSession(res.access_token, null);
        const me = await apiFetch<MeResponse>("/auth/me");
        if (!cancelled) setSession(res.access_token, me.role);
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
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!ready) {
    return null;
  }

  return <>{children}</>;
}
