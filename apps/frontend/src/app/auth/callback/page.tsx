"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { useAuthStore } from "@/store/auth";

export default function AuthCallbackPage() {
  const setAccessToken = useAuthStore((s) => s.setAccessToken);
  const router = useRouter();

  useEffect(() => {
    const params = new URLSearchParams(window.location.hash.slice(1));
    const token = params.get("access_token");
    const error = params.get("error");

    if (token) {
      setAccessToken(token);
      router.replace("/");
    } else {
      router.replace(`/login${error ? `?error=${error}` : ""}`);
    }
  }, [router, setAccessToken]);

  return (
    <main className="flex min-h-screen items-center justify-center">
      <p className="text-sm text-text-secondary">Signing you in…</p>
    </main>
  );
}
