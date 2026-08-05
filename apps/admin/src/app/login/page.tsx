"use client";

import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

interface AuthConfig {
  google_oauth_enabled: boolean;
  app_name: string;
}

interface LoginResponse {
  access_token: string;
}

// Sprint 14 sisipan (ADR 0012) — admin login is email+password by
// default now, must resolve to role Owner server-side (POST
// /auth/login/admin). Google stays available when /auth/config says
// it's enabled, same toggle as the main app.
export default function LoginPage() {
  const router = useRouter();
  const setSession = useAuthStore((s) => s.setSession);
  const [config, setConfig] = useState<AuthConfig | null>(null);
  const [email, setEmail] = useState("");
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
    if (!email || !password) return;
    setSubmitting(true);
    setError(null);
    try {
      const res = await apiFetch<LoginResponse>("/auth/login/admin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      // LoginAdmin only ever succeeds for role Owner (enforced
      // server-side) — safe to set directly without a round trip to
      // GET /auth/me.
      setSession(res.access_token, "owner");
      router.push("/");
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setError("Email atau password salah.");
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
        <h1 className="text-2xl font-bold">{config?.app_name ?? "Sonora"} Admin</h1>
        <p className="mt-2 text-sm text-text-secondary">Owner access only</p>
      </div>

      {config?.google_oauth_enabled && (
        <>
          <a
            href={`${API_BASE}/auth/google?app=admin`}
            className="flex w-full max-w-xs items-center justify-center gap-3 rounded-2xl bg-white px-6 py-3 font-semibold text-black"
          >
            Continue with Google
          </a>
          <div className="flex w-full max-w-xs items-center gap-3 text-xs text-text-secondary">
            <div className="h-px flex-1 bg-border" />
            or continue with email
            <div className="h-px flex-1 bg-border" />
          </div>
        </>
      )}

      <form onSubmit={handleSubmit} className="w-full max-w-xs space-y-3">
        <input
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          type="email"
          placeholder="Email"
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
    </main>
  );
}
