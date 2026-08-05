"use client";

import { Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";

interface AdminUser {
  id: string;
  name: string;
  email: string;
  role: "owner" | "member";
  status: "active" | "invited";
  created_at: string;
}

// Screens-spec #22 (Sprint 14 sisipan, ADR 0009). "Invite" doesn't send a
// real email — this project has no email infrastructure — it pre-creates
// a row that gets claimed automatically on that email's first real
// Google login. The Owner is expected to reach out out-of-band (e.g. the
// WhatsApp link on the user-facing Login page).
export default function UsersPage() {
  const [users, setUsers] = useState<AdminUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showInvite, setShowInvite] = useState(false);
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set());

  function load() {
    setLoading(true);
    apiFetch<AdminUser[]>("/admin/users")
      .then(setUsers)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load users"))
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  function removeAccess(id: string) {
    setBusyIds((s) => new Set(s).add(id));
    apiFetch(`/admin/users/${id}`, { method: "DELETE" })
      .then(() => setUsers((prev) => prev.filter((u) => u.id !== id)))
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to remove access"))
      .finally(() =>
        setBusyIds((s) => {
          const next = new Set(s);
          next.delete(id);
          return next;
        }),
      );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Users</h1>
          <p className="mt-1 text-sm text-text-secondary">Kelola akses Owner &amp; Member.</p>
        </div>
        <button
          onClick={() => setShowInvite(true)}
          className="flex items-center gap-2 rounded-control bg-primary px-4 py-2.5 text-sm font-semibold"
        >
          <Plus size={16} />
          Invite Member
        </button>
      </div>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      {showInvite && (
        <InviteModal
          onClose={() => setShowInvite(false)}
          onInvited={(user) => {
            setUsers((prev) => [...prev, user]);
            setShowInvite(false);
          }}
          onError={setError}
        />
      )}

      <div className="mt-8 overflow-x-auto rounded-card border border-border bg-card">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-text-secondary">
              <th className="px-4 py-3 font-medium">Nama</th>
              <th className="px-4 py-3 font-medium">Email</th>
              <th className="px-4 py-3 font-medium">Role</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Join</th>
              <th className="px-4 py-3 font-medium">Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-text-secondary">
                  Loading…
                </td>
              </tr>
            ) : (
              users.map((u) => (
                <tr key={u.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3">{u.name}</td>
                  <td className="px-4 py-3 text-text-secondary">{u.email}</td>
                  <td className="px-4 py-3 capitalize">{u.role}</td>
                  <td className="px-4 py-3">
                    <span
                      className={cn(
                        "rounded-full px-2.5 py-1 text-[11px] font-medium capitalize",
                        u.status === "active" ? "bg-success/15 text-success" : "bg-warning/15 text-warning",
                      )}
                    >
                      {u.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-text-secondary">{u.created_at}</td>
                  <td className="px-4 py-3">
                    {u.role !== "owner" && (
                      <button
                        onClick={() => removeAccess(u.id)}
                        disabled={busyIds.has(u.id)}
                        aria-label={`Remove access for ${u.email}`}
                        className="flex items-center justify-center rounded-control border border-border px-3 py-1.5 text-error hover:bg-error/10 disabled:opacity-50"
                      >
                        <Trash2 size={13} />
                      </button>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}

function InviteModal({
  onClose,
  onInvited,
  onError,
}: {
  onClose: () => void;
  onInvited: (user: AdminUser) => void;
  onError: (msg: string) => void;
}) {
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    apiFetch<AdminUser>("/admin/users/invite", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, name }),
    })
      .then(onInvited)
      .catch((e) => onError(e instanceof ApiError ? e.message : "Failed to invite user"))
      .finally(() => setSubmitting(false));
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <form onSubmit={submit} className="w-full max-w-sm rounded-card border border-border bg-card p-5">
        <h2 className="text-sm font-semibold">Invite Member</h2>
        <p className="mt-1 text-xs text-text-secondary">
          Tidak mengirim email — akses aktif otomatis begitu email ini login lewat Google pertama kali.
        </p>
        <input
          required
          type="email"
          placeholder="Email"
          value={email}
          onChange={(e) => setEmail(e.target.value)}
          className="mt-4 w-full rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <input
          placeholder="Nama (opsional)"
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="mt-3 w-full rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <div className="mt-4 flex gap-2">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 rounded-control border border-border px-4 py-2 text-sm font-medium text-text-secondary"
          >
            Batal
          </button>
          <button
            type="submit"
            disabled={submitting}
            className="flex-1 rounded-control bg-primary px-4 py-2 text-sm font-semibold disabled:opacity-50"
          >
            {submitting ? "Mengundang…" : "Invite"}
          </button>
        </div>
      </form>
    </div>
  );
}
