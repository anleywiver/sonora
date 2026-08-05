"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";
const OWNER_WHATSAPP = process.env.NEXT_PUBLIC_OWNER_WHATSAPP;
const REQUEST_ACCESS_MESSAGE =
  "Halo, saya mau minta akses ke Sonora. Apakah ada syarat tertentu?";

interface AuthConfig {
  google_oauth_enabled: boolean;
  app_name: string;
}

interface LoginResponse {
  access_token: string;
  device_id: string;
}

// Sprint 14 sisipan (ADR 0012) — credential login (username/password) is
// now the default path; Google stays fully wired but only shown when
// /auth/config says it's enabled (the backend also enforces this itself
// if someone hits /auth/google directly while it's off).
export default function LoginPage() {
  const router = useRouter();
  const setSession = useAuthStore((s) => s.setSession);
  const [config, setConfig] = useState<AuthConfig | null>(null);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    apiFetch<AuthConfig>("/auth/config")
      .then(setConfig)
      .catch(() => setConfig({ google_oauth_enabled: false, app_name: "Sonora" }));
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password) return;
    setSubmitting(true);
    setError(null);
    try {
      const res = await apiFetch<LoginResponse>("/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ username, password }),
      });
      setSession(res.access_token, res.device_id);
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("Username atau password salah.");
      } else {
        setError("Gagal login. Coba lagi.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 px-6">
      <div className="text-center">
        <div className="mx-auto mb-4 h-16 w-16 rounded-2xl bg-gradient-to-br from-accent to-primary" />
        <h1 className="text-2xl font-bold">{config?.app_name ?? "Sonora"}</h1>
        <p className="mt-2 text-sm text-text-secondary">Personal music streaming</p>
      </div>

      {config?.google_oauth_enabled && (
        <>
          <a
            href={`${API_BASE}/auth/google`}
            className="flex w-full max-w-xs items-center justify-center gap-3 rounded-2xl bg-white px-6 py-3 font-semibold text-black"
          >
            Continue with Google
          </a>
          {/* Apple sign-in has no backend (no Apple OAuth client
              configured) — shown disabled per screens-spec #2, "coming
              later" rather than omitted entirely. */}
          <button
            disabled
            className="flex w-full max-w-xs cursor-not-allowed items-center justify-center gap-3 rounded-2xl border border-border bg-white/5 px-6 py-3 font-semibold text-text-secondary opacity-50"
          >
            Continue with Apple
          </button>
          <div className="flex w-full max-w-xs items-center gap-3 text-xs text-text-secondary">
            <div className="h-px flex-1 bg-border" />
            or continue with username
            <div className="h-px flex-1 bg-border" />
          </div>
        </>
      )}

      <form onSubmit={handleSubmit} className="w-full max-w-xs space-y-3">
        <input
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="Username"
          autoComplete="username"
          className="w-full rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm outline-none placeholder:text-text-secondary"
        />
        <input
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          type="password"
          placeholder="Password"
          autoComplete="current-password"
          className="w-full rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm outline-none placeholder:text-text-secondary"
        />
        {error && <p className="text-xs text-error">{error}</p>}
        <button
          type="submit"
          disabled={submitting}
          className="w-full rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white disabled:opacity-50"
        >
          {submitting ? "Signing in..." : "Sign in"}
        </button>
      </form>

      {/* Sonora is invite-only — this is a contact shortcut, not a
          self-signup mechanism. No registration form/endpoint exists or
          should exist here; access is still granted entirely by the
          Owner out-of-band. */}
      {OWNER_WHATSAPP && (
        <a
          href={`https://wa.me/${OWNER_WHATSAPP}?text=${encodeURIComponent(REQUEST_ACCESS_MESSAGE)}`}
          target="_blank"
          rel="noopener noreferrer"
          className="text-xs text-text-secondary underline underline-offset-2"
        >
          Belum punya akses? Request akses
        </a>
      )}
    </main>
  );
}
