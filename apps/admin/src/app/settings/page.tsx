"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

interface Settings {
  app_name: string;
  default_language: string;
  google_oauth_enabled: string;
  maintenance_mode: string;
}

// Sprint 14 sisipan (ADR 0012) — app_name/default_language were the
// original Admin Settings scope; google_oauth_enabled was folded in
// later once the credential-auth pivot needed a runtime toggle, same
// key-value store. Storage account settings live in Drive Manager, not
// duplicated here (link only).
export default function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState<string | null>(null);
  const [appName, setAppName] = useState("");

  function load() {
    apiFetch<Settings>("/admin/settings")
      .then((s) => {
        setSettings(s);
        setAppName(s.app_name);
      })
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load settings"));
  }

  useEffect(load, []);

  async function update(key: string, value: string) {
    setSaving(key);
    try {
      await apiFetch("/admin/settings", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ key, value }),
      });
      setSettings((s) => (s ? { ...s, [key]: value } : s));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "Failed to update setting");
    } finally {
      setSaving(null);
    }
  }

  if (!settings) {
    return (
      <main className="p-8">
        <h1 className="text-xl font-bold">Settings</h1>
        {error ? <p className="mt-4 text-sm text-error">{error}</p> : <p className="mt-4 text-sm text-text-secondary">Memuat...</p>}
      </main>
    );
  }

  return (
    <main className="p-8">
      <h1 className="text-xl font-bold">Settings</h1>
      {error && <p className="mt-2 text-sm text-error">{error}</p>}

      <section className="mt-6 max-w-md rounded-[20px] border border-border bg-white/5 p-5">
        <h2 className="text-sm font-semibold">App Name</h2>
        <div className="mt-3 flex gap-2">
          <input
            value={appName}
            onChange={(e) => setAppName(e.target.value)}
            className="flex-1 rounded-2xl border border-border bg-white/5 px-4 py-2 text-sm outline-none"
          />
          <button
            onClick={() => update("app_name", appName)}
            disabled={saving === "app_name"}
            className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
          >
            Save
          </button>
        </div>
      </section>

      <section className="mt-4 max-w-md rounded-[20px] border border-border bg-white/5 p-5">
        <h2 className="text-sm font-semibold">Default Language</h2>
        <div className="mt-3 flex gap-2">
          {(["id", "en"] as const).map((lang) => (
            <button
              key={lang}
              onClick={() => update("default_language", lang)}
              disabled={saving === "default_language"}
              className={cn(
                "rounded-2xl border border-border px-4 py-2 text-sm",
                settings.default_language === lang ? "bg-primary text-white" : "text-text-secondary",
              )}
            >
              {lang === "id" ? "Indonesia" : "English"}
            </button>
          ))}
        </div>
      </section>

      <ToggleCard
        title="Enable Google OAuth Login"
        description="Kalau dimatikan, tombol Continue with Google disembunyikan di halaman login (user & admin) DAN backend menolak /auth/google langsung dengan 403 — bukan cuma disembunyikan di UI."
        enabled={settings.google_oauth_enabled === "true"}
        busy={saving === "google_oauth_enabled"}
        onToggle={(next) => update("google_oauth_enabled", next ? "true" : "false")}
      />

      <ToggleCard
        title="Maintenance Mode"
        description="Kalau dinyalakan, semua endpoint API non-Owner menolak request dengan 503. Login dan halaman admin tetap jalan supaya Owner bisa mematikannya lagi."
        enabled={settings.maintenance_mode === "true"}
        busy={saving === "maintenance_mode"}
        onToggle={(next) => update("maintenance_mode", next ? "true" : "false")}
      />

      <section className="mt-4 max-w-md rounded-[20px] border border-border bg-white/5 p-5">
        <h2 className="text-sm font-semibold">Storage</h2>
        <p className="mt-2 text-xs text-text-secondary">
          Pengaturan Google Drive storage pool ada di halaman terpisah.
        </p>
        <Link href="/drive-manager" className="mt-3 inline-block text-sm text-hover underline underline-offset-2">
          Buka Drive Manager
        </Link>
      </section>
    </main>
  );
}

function ToggleCard({
  title,
  description,
  enabled,
  busy,
  onToggle,
}: {
  title: string;
  description: string;
  enabled: boolean;
  busy: boolean;
  onToggle: (next: boolean) => void;
}) {
  return (
    <section className="mt-4 max-w-md rounded-[20px] border border-border bg-white/5 p-5">
      <div className="flex items-center justify-between gap-4">
        <h2 className="text-sm font-semibold">{title}</h2>
        <button
          onClick={() => onToggle(!enabled)}
          disabled={busy}
          aria-pressed={enabled}
          className={cn(
            "relative h-6 w-11 flex-shrink-0 rounded-full transition-colors disabled:opacity-50",
            enabled ? "bg-primary" : "bg-white/10",
          )}
        >
          <span
            className={cn(
              "absolute top-0.5 h-5 w-5 rounded-full bg-white transition-transform",
              enabled ? "translate-x-5" : "translate-x-0.5",
            )}
          />
        </button>
      </div>
      <p className="mt-2 text-xs text-text-secondary">{description}</p>
    </section>
  );
}
