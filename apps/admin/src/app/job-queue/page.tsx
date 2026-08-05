"use client";

import { RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { cn } from "@/lib/utils";
import { useAuthStore } from "@/store/auth";

interface Job {
  id: string;
  user_id: string;
  source_type: string;
  status: "pending" | "processing" | "completed" | "failed";
  error_message: string | null;
  created_at: string;
}

interface JobListEnvelope {
  data: Job[];
  next_cursor: string;
  has_more: boolean;
}

const statusBadge: Record<Job["status"], string> = {
  pending: "bg-white/5 text-text-secondary",
  processing: "bg-warning/15 text-warning",
  completed: "bg-success/15 text-success",
  failed: "bg-error/15 text-error",
};

const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

// GET /admin/jobs returns {data, next_cursor, has_more} — apiFetch always
// unwraps to just `data`, which would drop the pagination fields this
// page doesn't currently use but the envelope still carries.
async function fetchJobs(): Promise<JobListEnvelope> {
  const token = useAuthStore.getState().accessToken;
  const res = await fetch(`${API_BASE}/admin/jobs?limit=50`, {
    credentials: "include",
    headers: token ? { Authorization: `Bearer ${token}` } : {},
  });
  const body = await res.json();
  if (!res.ok) {
    throw new ApiError(res.status, body?.error?.code ?? "unknown", body?.error?.message ?? "Request failed");
  }
  return body;
}

export default function JobQueuePage() {
  const [jobs, setJobs] = useState<Job[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [retryingIds, setRetryingIds] = useState<Set<string>>(new Set());

  function load() {
    setLoading(true);
    fetchJobs()
      .then((res) => setJobs(res.data))
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load jobs"))
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  function retry(id: string) {
    setRetryingIds((s) => new Set(s).add(id));
    apiFetch(`/admin/jobs/${id}/retry`, { method: "POST" })
      .then(load)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Retry failed"))
      .finally(() =>
        setRetryingIds((s) => {
          const next = new Set(s);
          next.delete(id);
          return next;
        }),
      );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Job Queue</h1>
      <p className="mt-1 text-sm text-text-secondary">
        Semua ingest job dari semua user — `ingest_jobs` table.
      </p>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      <div className="mt-8 overflow-x-auto rounded-card border border-border bg-card">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-text-secondary">
              <th className="px-4 py-3 font-medium">Job</th>
              <th className="px-4 py-3 font-medium">Type</th>
              <th className="px-4 py-3 font-medium">Status</th>
              <th className="px-4 py-3 font-medium">Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-text-secondary">
                  Loading…
                </td>
              </tr>
            ) : jobs.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-6 text-center text-text-secondary">
                  Belum ada ingest job.
                </td>
              </tr>
            ) : (
              jobs.map((job) => {
                const retrying = retryingIds.has(job.id);
                return (
                  <tr key={job.id} className="border-b border-border last:border-0">
                    <td className="px-4 py-3 font-mono text-xs">{job.id.slice(0, 8)}</td>
                    <td className="px-4 py-3">{job.source_type}</td>
                    <td className="px-4 py-3">
                      <span className={cn("rounded-full px-2.5 py-1 text-[11px] font-medium", statusBadge[job.status])}>
                        {job.status}
                      </span>
                      {job.error_message && (
                        <p className="mt-1 max-w-xs truncate text-xs text-error" title={job.error_message}>
                          {job.error_message}
                        </p>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {job.status === "failed" && (
                        <button
                          onClick={() => retry(job.id)}
                          disabled={retrying}
                          className="flex items-center gap-1.5 rounded-control border border-border px-3 py-1.5 text-xs font-medium text-text-secondary hover:text-text-primary disabled:opacity-50"
                        >
                          <RefreshCw size={12} className={retrying ? "animate-spin" : ""} />
                          Retry
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    </main>
  );
}
