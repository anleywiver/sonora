"use client";

import { Plus, RefreshCw, Trash2, X } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

interface Connection {
  id: string;
  provider: "bandcamp" | "cloud_sync";
  label: string;
  account_email: string;
  is_active: boolean;
  last_synced_at: string | null;
}

interface FilterRule {
  id: string;
  source_type: string;
  rule_type: "genre_allow" | "year_min" | "year_max";
  value: string;
}

const providerLabel: Record<Connection["provider"], string> = {
  bandcamp: "Bandcamp",
  cloud_sync: "Cloud Sync (Dropbox)",
};

// Manual Upload is always available (it's just the /ingest/upload
// endpoint, no connection row) — only Bandcamp/cloud sync are real
// ingest_source_connections (Sprint 10, ADR 0004).
export default function IngestSourcesPage() {
  const [connections, setConnections] = useState<Connection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set());
  const [showForm, setShowForm] = useState(false);

  function load() {
    setLoading(true);
    apiFetch<Connection[]>("/admin/ingest-sources/connections")
      .then(setConnections)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load"))
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  function withBusy(id: string, fn: () => Promise<void>) {
    setBusyIds((s) => new Set(s).add(id));
    fn().finally(() =>
      setBusyIds((s) => {
        const next = new Set(s);
        next.delete(id);
        return next;
      }),
    );
  }

  function syncNow(id: string) {
    setError(null);
    withBusy(id, () =>
      apiFetch(`/admin/ingest-sources/connections/${id}/sync`, { method: "POST" })
        .then(load)
        .catch((e) => setError(e instanceof ApiError ? e.message : "Sync failed")),
    );
  }

  function disconnect(id: string) {
    withBusy(id, () =>
      apiFetch(`/admin/ingest-sources/connections/${id}`, { method: "DELETE" })
        .then(() => setConnections((prev) => prev.filter((c) => c.id !== id)))
        .catch((e) => setError(e instanceof ApiError ? e.message : "Disconnect failed")),
    );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Ingest Sources</h1>
          <p className="mt-1 text-sm text-text-secondary">
            Sumber ingest legal — manual upload, Bandcamp, cloud sync. Tidak ada scraping/auto-download.
          </p>
        </div>
        <button
          onClick={() => setShowForm((v) => !v)}
          className="flex items-center gap-2 rounded-control bg-primary px-4 py-2.5 text-sm font-semibold"
        >
          <Plus size={16} />
          Connect provider
        </button>
      </div>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      {showForm && (
        <ConnectForm
          onConnected={(conn) => {
            setConnections((prev) => [...prev, conn]);
            setShowForm(false);
          }}
          onError={setError}
        />
      )}

      <div className="mt-8 space-y-3">
        <div className="flex items-center justify-between rounded-card border border-border bg-card p-4">
          <span className="font-medium">Manual Upload</span>
          <span className="rounded-full bg-success/15 px-2.5 py-1 text-[11px] font-medium capitalize text-success">
            connected
          </span>
        </div>

        {loading ? (
          <p className="text-sm text-text-secondary">Loading…</p>
        ) : (
          connections.map((conn) => {
            const busy = busyIds.has(conn.id);
            return (
              <div
                key={conn.id}
                className="flex items-center justify-between rounded-card border border-border bg-card p-4"
              >
                <div>
                  <p className="font-medium">
                    {conn.label}{" "}
                    <span className="text-xs font-normal text-text-secondary">
                      ({providerLabel[conn.provider]})
                    </span>
                  </p>
                  <p className="text-xs text-text-secondary">
                    {conn.last_synced_at
                      ? `Last synced ${new Date(conn.last_synced_at).toLocaleString("id-ID")}`
                      : "Never synced"}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span
                    className={cn(
                      "rounded-full px-2.5 py-1 text-[11px] font-medium capitalize",
                      conn.is_active ? "bg-success/15 text-success" : "bg-white/5 text-text-secondary",
                    )}
                  >
                    {conn.is_active ? "connected" : "disconnected"}
                  </span>
                  <button
                    onClick={() => syncNow(conn.id)}
                    disabled={busy}
                    className="flex items-center justify-center gap-1.5 rounded-control border border-border px-3 py-2 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-50"
                  >
                    <RefreshCw size={13} className={busy ? "animate-spin" : ""} />
                    Sync now
                  </button>
                  <button
                    onClick={() => disconnect(conn.id)}
                    disabled={busy}
                    aria-label="Disconnect"
                    className="flex items-center justify-center rounded-control border border-border px-3 py-2 text-error hover:bg-error/10 disabled:opacity-50"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            );
          })
        )}
      </div>

      <div className="mt-8 grid grid-cols-1 gap-4 lg:grid-cols-2">
        <FilterRulesPanel sourceType="bandcamp" title="Bandcamp Filter Rules" />
        <FilterRulesPanel sourceType="cloud_sync" title="Cloud Sync Filter Rules" />
      </div>
    </main>
  );
}

// Sprint 14 sisipan (ADR 0008) — genre allow-list + year range, applies
// ONLY to auto-ingest from this source. Manual upload is never filtered,
// so that's called out explicitly rather than left implicit.
function FilterRulesPanel({ sourceType, title }: { sourceType: "bandcamp" | "cloud_sync"; title: string }) {
  const [rules, setRules] = useState<FilterRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [genreInput, setGenreInput] = useState("");
  const [yearMinInput, setYearMinInput] = useState("");
  const [yearMaxInput, setYearMaxInput] = useState("");

  function load() {
    setLoading(true);
    apiFetch<FilterRule[]>(`/admin/ingest-sources/${sourceType}/filters`)
      .then(setRules)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load filter rules"))
      .finally(() => setLoading(false));
  }

  useEffect(load, [sourceType]);

  function addRule(ruleType: FilterRule["rule_type"], value: string) {
    if (!value.trim()) return;
    apiFetch(`/admin/ingest-sources/${sourceType}/filters`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ rule_type: ruleType, value: value.trim() }),
    })
      .then(load)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to add rule"));
  }

  function removeRule(id: string) {
    apiFetch(`/admin/ingest-sources/${sourceType}/filters/${id}`, { method: "DELETE" })
      .then(load)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to remove rule"));
  }

  const genreRules = rules.filter((r) => r.rule_type === "genre_allow");
  const yearMin = rules.find((r) => r.rule_type === "year_min");
  const yearMax = rules.find((r) => r.rule_type === "year_max");

  return (
    <div className="rounded-card border border-border bg-card p-5">
      <h2 className="text-sm font-semibold text-text-secondary">{title}</h2>
      <p className="mt-1 text-xs text-text-secondary">
        Cuma berlaku untuk auto-ingest dari {sourceType === "bandcamp" ? "Bandcamp" : "cloud sync"} — manual upload
        tidak pernah difilter.
      </p>

      {error && <p className="mt-2 text-xs text-error">{error}</p>}

      {loading ? (
        <p className="mt-4 text-sm text-text-secondary">Loading…</p>
      ) : (
        <>
          <div className="mt-4">
            <p className="mb-2 text-xs font-medium text-text-secondary">Genre allow-list</p>
            <div className="flex flex-wrap gap-2">
              {genreRules.map((r) => (
                <span
                  key={r.id}
                  className="flex items-center gap-1.5 rounded-full bg-accent/15 px-3 py-1 text-xs text-accent"
                >
                  {r.value}
                  <button onClick={() => removeRule(r.id)} aria-label={`Remove ${r.value}`}>
                    <X size={12} />
                  </button>
                </span>
              ))}
              {genreRules.length === 0 && (
                <span className="text-xs text-text-secondary">Semua genre diperbolehkan (belum ada rule)</span>
              )}
            </div>
            <div className="mt-2 flex gap-2">
              <input
                value={genreInput}
                onChange={(e) => setGenreInput(e.target.value)}
                placeholder="Tambah genre (mis. Jazz)"
                className="flex-1 rounded-control border border-border bg-background px-3 py-1.5 text-xs"
              />
              <button
                onClick={() => {
                  addRule("genre_allow", genreInput);
                  setGenreInput("");
                }}
                className="rounded-control bg-primary px-3 py-1.5 text-xs font-semibold"
              >
                <Plus size={12} />
              </button>
            </div>
          </div>

          <div className="mt-4">
            <p className="mb-2 text-xs font-medium text-text-secondary">Year range</p>
            <div className="flex items-center gap-2">
              {yearMin ? (
                <span className="flex items-center gap-1.5 rounded-full bg-accent/15 px-3 py-1 text-xs text-accent">
                  Min {yearMin.value}
                  <button onClick={() => removeRule(yearMin.id)} aria-label="Remove year min">
                    <X size={12} />
                  </button>
                </span>
              ) : (
                <input
                  value={yearMinInput}
                  onChange={(e) => setYearMinInput(e.target.value)}
                  placeholder="Min tahun"
                  className="w-28 rounded-control border border-border bg-background px-3 py-1.5 text-xs"
                />
              )}
              {!yearMin && (
                <button
                  onClick={() => {
                    addRule("year_min", yearMinInput);
                    setYearMinInput("");
                  }}
                  className="rounded-control bg-primary px-3 py-1.5 text-xs font-semibold"
                >
                  <Plus size={12} />
                </button>
              )}
              {yearMax ? (
                <span className="flex items-center gap-1.5 rounded-full bg-accent/15 px-3 py-1 text-xs text-accent">
                  Max {yearMax.value}
                  <button onClick={() => removeRule(yearMax.id)} aria-label="Remove year max">
                    <X size={12} />
                  </button>
                </span>
              ) : (
                <input
                  value={yearMaxInput}
                  onChange={(e) => setYearMaxInput(e.target.value)}
                  placeholder="Max tahun"
                  className="w-28 rounded-control border border-border bg-background px-3 py-1.5 text-xs"
                />
              )}
              {!yearMax && (
                <button
                  onClick={() => {
                    addRule("year_max", yearMaxInput);
                    setYearMaxInput("");
                  }}
                  className="rounded-control bg-primary px-3 py-1.5 text-xs font-semibold"
                >
                  <Plus size={12} />
                </button>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}

function ConnectForm({
  onConnected,
  onError,
}: {
  onConnected: (conn: Connection) => void;
  onError: (msg: string) => void;
}) {
  const [provider, setProvider] = useState<Connection["provider"]>("bandcamp");
  const [label, setLabel] = useState("");
  const [accountEmail, setAccountEmail] = useState("");
  const [identityCookie, setIdentityCookie] = useState("");
  const [fanId, setFanId] = useState("");
  const [refreshToken, setRefreshToken] = useState("");
  const [appKey, setAppKey] = useState("");
  const [appSecret, setAppSecret] = useState("");
  const [folderPath, setFolderPath] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    const body =
      provider === "bandcamp"
        ? { provider, label, account_email: accountEmail, identity_cookie: identityCookie, fan_id: fanId }
        : {
            provider,
            label,
            account_email: accountEmail,
            refresh_token: refreshToken,
            app_key: appKey,
            app_secret: appSecret,
            folder_path: folderPath,
          };

    apiFetch<Connection>("/admin/ingest-sources/connections", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    })
      .then(onConnected)
      .catch((e) => onError(e instanceof ApiError ? e.message : "Failed to connect"))
      .finally(() => setSubmitting(false));
  }

  return (
    <form onSubmit={submit} className="mt-6 rounded-card border border-border bg-card p-5">
      <div className="mb-4 flex gap-2">
        {(["bandcamp", "cloud_sync"] as const).map((p) => (
          <button
            key={p}
            type="button"
            onClick={() => setProvider(p)}
            className={cn(
              "rounded-control px-3 py-1.5 text-xs font-medium",
              provider === p ? "bg-primary text-white" : "border border-border text-text-secondary",
            )}
          >
            {providerLabel[p]}
          </button>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
        <input
          required
          placeholder="Label"
          value={label}
          onChange={(e) => setLabel(e.target.value)}
          className="rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <input
          placeholder="Account email"
          value={accountEmail}
          onChange={(e) => setAccountEmail(e.target.value)}
          className="rounded-control border border-border bg-background px-3 py-2 text-sm"
        />

        {provider === "bandcamp" ? (
          <>
            <input
              required
              placeholder="Identity cookie (dari browser, login bandcamp.com)"
              value={identityCookie}
              onChange={(e) => setIdentityCookie(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm md:col-span-2"
            />
            <input
              required
              placeholder="Fan ID"
              value={fanId}
              onChange={(e) => setFanId(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm"
            />
          </>
        ) : (
          <>
            <input
              required
              placeholder="Refresh token (Dropbox OAuth)"
              value={refreshToken}
              onChange={(e) => setRefreshToken(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm md:col-span-2"
            />
            <input
              required
              placeholder="App key"
              value={appKey}
              onChange={(e) => setAppKey(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm"
            />
            <input
              required
              placeholder="App secret"
              value={appSecret}
              onChange={(e) => setAppSecret(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm"
            />
            <input
              required
              placeholder="Folder path (mis. /Sonora Sync)"
              value={folderPath}
              onChange={(e) => setFolderPath(e.target.value)}
              className="rounded-control border border-border bg-background px-3 py-2 text-sm md:col-span-2"
            />
          </>
        )}
      </div>

      <button
        type="submit"
        disabled={submitting}
        className="mt-4 rounded-control bg-primary px-4 py-2 text-sm font-semibold disabled:opacity-50"
      >
        {submitting ? "Menghubungkan…" : "Connect"}
      </button>
    </form>
  );
}
