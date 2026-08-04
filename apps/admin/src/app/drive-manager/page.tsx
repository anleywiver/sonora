"use client";

import { Plus, RefreshCw, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn, formatBytes } from "@/lib/utils";

interface StorageAccount {
  id: string;
  provider: string;
  label: string;
  account_email: string;
  quota_bytes: number | null;
  used_bytes: number;
  is_active: boolean;
  health_status: "healthy" | "degraded" | "down";
  last_health_check_at: string | null;
}

const healthBadge: Record<StorageAccount["health_status"], string> = {
  healthy: "bg-success/15 text-success",
  degraded: "bg-warning/15 text-warning",
  down: "bg-error/15 text-error",
};

export default function DriveManagerPage() {
  const [accounts, setAccounts] = useState<StorageAccount[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set());
  const [showForm, setShowForm] = useState(false);

  function load() {
    setLoading(true);
    apiFetch<StorageAccount[]>("/admin/storage/accounts")
      .then(setAccounts)
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

  function runHealthCheck(id: string) {
    withBusy(id, () =>
      apiFetch<StorageAccount>(`/admin/storage/accounts/${id}/health-check`, { method: "POST" })
        .then((updated) => setAccounts((prev) => prev.map((a) => (a.id === id ? updated : a))))
        .catch((e) => setError(e instanceof ApiError ? e.message : "Health check failed")),
    );
  }

  function deleteAccount(id: string) {
    withBusy(id, () =>
      apiFetch(`/admin/storage/accounts/${id}`, { method: "DELETE" })
        .then(() => setAccounts((prev) => prev.filter((a) => a.id !== id)))
        .catch((e) => setError(e instanceof ApiError ? e.message : "Delete failed")),
    );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Drive Manager</h1>
          <p className="mt-1 text-sm text-text-secondary">
            Storage account Google Drive untuk ingest pipeline (quota-aware routing).
          </p>
        </div>
        <button
          onClick={() => setShowForm((v) => !v)}
          className="flex items-center gap-2 rounded-control bg-primary px-4 py-2.5 text-sm font-semibold"
        >
          <Plus size={16} />
          Add drive
        </button>
      </div>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      {showForm && (
        <AddDriveForm
          onCreated={(account) => {
            setAccounts((prev) => [...prev, account]);
            setShowForm(false);
          }}
          onError={(msg) => setError(msg)}
        />
      )}

      {loading ? (
        <p className="mt-8 text-sm text-text-secondary">Loading…</p>
      ) : accounts.length === 0 ? (
        <p className="mt-8 text-sm text-text-secondary">Belum ada storage account.</p>
      ) : (
        <div className="mt-8 grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {accounts.map((account) => {
            const busy = busyIds.has(account.id);
            const usedPct = account.quota_bytes
              ? Math.min(100, (account.used_bytes / account.quota_bytes) * 100)
              : 0;
            const remaining = account.quota_bytes
              ? account.quota_bytes - account.used_bytes
              : null;

            return (
              <div key={account.id} className="rounded-card border border-border bg-card p-5">
                <div className="flex items-start justify-between">
                  <div>
                    <p className="font-semibold">{account.label}</p>
                    <p className="text-xs text-text-secondary">{account.account_email}</p>
                  </div>
                  <span
                    className={cn(
                      "rounded-full px-2.5 py-1 text-[11px] font-medium capitalize",
                      healthBadge[account.health_status],
                    )}
                  >
                    {account.health_status}
                  </span>
                </div>

                <div className="mt-4">
                  <div className="h-1.5 w-full overflow-hidden rounded-full bg-white/10">
                    <div
                      className={cn(
                        "h-full rounded-full",
                        usedPct > 90 ? "bg-error" : "bg-accent",
                      )}
                      style={{ width: `${usedPct}%` }}
                    />
                  </div>
                  <p className="mt-2 text-xs text-text-secondary">
                    {formatBytes(account.used_bytes)} used ·{" "}
                    {remaining === null ? "Unlimited" : `${formatBytes(remaining)} left`}
                  </p>
                </div>

                <div className="mt-4 flex gap-2">
                  <button
                    onClick={() => runHealthCheck(account.id)}
                    disabled={busy}
                    className="flex flex-1 items-center justify-center gap-1.5 rounded-control border border-border px-3 py-2 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-50"
                  >
                    <RefreshCw size={13} className={busy ? "animate-spin" : ""} />
                    Run health check
                  </button>
                  <button
                    onClick={() => deleteAccount(account.id)}
                    disabled={busy}
                    aria-label="Delete drive"
                    className="flex items-center justify-center rounded-control border border-border px-3 py-2 text-error hover:bg-error/10 disabled:opacity-50"
                  >
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            );
          })}
        </div>
      )}
    </main>
  );
}

function AddDriveForm({
  onCreated,
  onError,
}: {
  onCreated: (account: StorageAccount) => void;
  onError: (msg: string) => void;
}) {
  const [label, setLabel] = useState("");
  const [accountEmail, setAccountEmail] = useState("");
  const [refreshToken, setRefreshToken] = useState("");
  const [quotaBytes, setQuotaBytes] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    apiFetch<StorageAccount>("/admin/storage/accounts", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        label,
        account_email: accountEmail,
        refresh_token: refreshToken,
        quota_bytes: quotaBytes ? Number(quotaBytes) : null,
      }),
    })
      .then(onCreated)
      .catch((e) => onError(e instanceof ApiError ? e.message : "Failed to create"))
      .finally(() => setSubmitting(false));
  }

  return (
    <form
      onSubmit={submit}
      className="mt-6 grid grid-cols-1 gap-3 rounded-card border border-border bg-card p-5 md:grid-cols-2"
    >
      <input
        required
        placeholder="Label (mis. Drive Utama)"
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
      <input
        required
        placeholder="Refresh token (dari OAuth consent manual)"
        value={refreshToken}
        onChange={(e) => setRefreshToken(e.target.value)}
        className="rounded-control border border-border bg-background px-3 py-2 text-sm md:col-span-2"
      />
      <input
        placeholder="Quota bytes (kosongkan = unlimited)"
        value={quotaBytes}
        onChange={(e) => setQuotaBytes(e.target.value)}
        className="rounded-control border border-border bg-background px-3 py-2 text-sm"
      />
      <button
        type="submit"
        disabled={submitting}
        className="rounded-control bg-primary px-4 py-2 text-sm font-semibold disabled:opacity-50 md:col-span-2"
      >
        {submitting ? "Menyimpan…" : "Simpan"}
      </button>
    </form>
  );
}
