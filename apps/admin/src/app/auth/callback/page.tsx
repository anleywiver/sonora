"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

interface MeResponse {
  role: string;
}

export default function AuthCallbackPage() {
  const setSession = useAuthStore((s) => s.setSession);
  const router = useRouter();

  useEffect(() => {
    const params = new URLSearchParams(window.location.hash.slice(1));
    const token = params.get("access_token");
    const error = params.get("error");

    if (!token) {
      router.replace(`/login${error ? `?error=${error}` : ""}`);
      return;
    }

    setSession(token, null);
    apiFetch<MeResponse>("/auth/me")
      .then((me) => {
        setSession(token, me.role);
        router.replace("/");
      })
      .catch((e) => {
        setSession(null, null);
        router.replace(`/login?error=${e instanceof ApiError ? e.code : "unknown"}`);
      });
  }, [router, setSession]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-text-secondary">Signing you in…</p>
    </main>
  );
}
