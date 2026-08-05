"use client";

import { ChevronRight, Trash2 } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { deleteDownload, listDownloads } from "@/lib/offline-db";

interface Me {
  role: string;
}

interface StorageAccount {
  id: string;
}

function formatSize(bytes: number): string {
  if (bytes === 0) return "0 MB";
  const mb = bytes / (1024 * 1024);
  return mb >= 1024 ? `${(mb / 1024).toFixed(1)} GB` : `${mb.toFixed(1)} MB`;
}

// ui-implementation-spec.md #6.3 — built out for real this sprint (was a
// bare "Profile" link + placeholder text before). "Connected Google
// Drive" from the spec's mockup only renders for Owner: that's admin-
// scoped data (GET /admin/storage/accounts, RequireRole(owner)) a Member
// can't and shouldn't see — shown conditionally rather than faked or
// exposing a new endpoint just for this cosmetic row.
export default function SettingsPage() {
  const [role, setRole] = useState<string | null>(null);
  const [driveCount, setDriveCount] = useState<number | null>(null);
  const [downloadBytes, setDownloadBytes] = useState(0);
  const [downloadCount, setDownloadCount] = useState(0);
  const [clearing, setClearing] = useState(false);

  useEffect(() => {
    apiFetch<Me>("/auth/me").then((me) => {
      setRole(me.role);
      if (me.role === "owner") {
        apiFetch<StorageAccount[]>("/admin/storage/accounts")
          .then((accounts) => setDriveCount(accounts.length))
          .catch(() => setDriveCount(null));
      }
    });
    loadDownloads();
  }, []);

  function loadDownloads() {
    listDownloads().then((rows) => {
      setDownloadCount(rows.length);
      setDownloadBytes(rows.reduce((sum, r) => sum + r.sizeBytes, 0));
    });
  }

  async function clearDownloads() {
    setClearing(true);
    const rows = await listDownloads();
    await Promise.all(rows.map((r) => deleteDownload(r.songId)));
    loadDownloads();
    setClearing(false);
  }

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <h1 className="text-xl font-bold">Settings</h1>

      <div className="mt-6 space-y-6">
        <section>
          <p className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
            Appearance
          </p>
          <div className="flex items-center justify-between border-b border-border py-2.5">
            <span className="text-xs font-medium">Theme</span>
            <span className="text-[11px] text-text-secondary">Dark</span>
          </div>
        </section>

        <section>
          <p className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
            Storage
          </p>
          <div className="py-2.5">
            <div className="mb-1.5 flex justify-between">
              <span className="text-xs font-medium">Musik terdownload</span>
              <span className="text-[11px] text-text-secondary">
                {downloadCount} lagu · {formatSize(downloadBytes)}
              </span>
            </div>
          </div>
          {role === "owner" && (
            <div className="flex items-center justify-between border-b border-border py-2.5">
              <div>
                <p className="text-xs font-medium">Connected Google Drive</p>
                <p className="mt-0.5 text-[10px] text-text-secondary">
                  {driveCount ?? "…"} account{driveCount === 1 ? "" : "s"} linked
                </p>
              </div>
              <ChevronRight size={14} className="text-text-secondary" />
            </div>
          )}
          <button
            onClick={clearDownloads}
            disabled={clearing || downloadCount === 0}
            className="flex w-full items-center justify-between border-b border-border py-2.5 text-left disabled:opacity-50"
          >
            <span className="flex items-center gap-1.5 text-xs font-medium text-error">
              <Trash2 size={12} />
              Clear downloaded songs
            </span>
            <span className="text-[11px] text-text-secondary">{formatSize(downloadBytes)}</span>
          </button>
          <Link href="/downloads" className="mt-2 inline-block text-[11px] text-hover underline underline-offset-2">
            Kelola downloads
          </Link>
        </section>

        <section>
          <p className="mb-2 text-[10px] font-semibold uppercase tracking-wide text-text-secondary">
            Playback
          </p>
          <div className="flex items-center justify-between border-b border-border py-2.5">
            <span className="text-xs font-medium">Lyrics source</span>
            <span className="flex items-center gap-1 text-[11px] text-text-secondary">
              LRCLIB (auto)
            </span>
          </div>
        </section>

        <div className="flex items-center justify-between py-2.5">
          <span className="text-xs font-medium">About</span>
          <span className="text-[11px] text-text-secondary">Sonora v1.0.0</span>
        </div>

        <Link
          href="/profile"
          className="block w-full rounded-2xl bg-primary px-6 py-3 text-center text-sm font-semibold text-white"
        >
          Profile
        </Link>
      </div>
    </main>
  );
}
