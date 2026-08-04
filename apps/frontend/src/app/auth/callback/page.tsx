"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuthStore } from "@/store/auth";
import { useWSStore } from "@/store/ws";

export default function AuthCallbackPage() {
  const setSession = useAuthStore((s) => s.setSession);
  const router = useRouter();

  useEffect(() => {
    const params = new URLSearchParams(window.location.hash.slice(1));
    const token = params.get("access_token");
    const deviceId = params.get("device_id");
    const error = params.get("error");

    if (token && deviceId) {
      setSession(token, deviceId);
      void useWSStore.getState().connect();
      router.replace("/");
    } else {
      router.replace(`/login${error ? `?error=${error}` : ""}`);
    }
  }, [router, setSession]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-text-secondary">Signing you in…</p>
    </main>
  );
}
