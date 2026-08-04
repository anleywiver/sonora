"use client";

import { Trash2 } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { deleteDownload, listDownloads } from "@/lib/offline-db";

interface DownloadItem {
  songId: string;
  title: string;
  artistName: string;
  sizeBytes: number;
}

function formatSize(bytes: number): string {
  const mb = bytes / (1024 * 1024);
  return `${mb.toFixed(1)} MB`;
}

export default function DownloadsPage() {
  const [items, setItems] = useState<DownloadItem[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    listDownloads()
      .then((rows) =>
        setItems(
          rows.map((r) => ({
            songId: r.songId,
            title: r.title,
            artistName: r.artistName,
            sizeBytes: r.sizeBytes,
          })),
        ),
      )
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleDelete = async (songId: string) => {
    await deleteDownload(songId);
    load();
  };

  const totalBytes = items.reduce((sum, i) => sum + i.sizeBytes, 0);

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <h1 className="text-xl font-bold">Downloads</h1>
      {items.length > 0 && (
        <p className="mt-1 text-xs text-text-secondary">
          {items.length} lagu · {formatSize(totalBytes)} tersimpan di perangkat ini
        </p>
      )}

      {loading && <p className="mt-6 text-sm text-text-secondary">Memuat...</p>}

      <ul className="mt-4 space-y-2">
        {items.map((item) => (
          <li
            key={item.songId}
            className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5"
          >
            <Link href={`/song/${item.songId}`} className="flex min-w-0 flex-1 items-center gap-3">
              <div className="h-12 w-12 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
              <div className="min-w-0">
                <p className="truncate text-sm font-medium">{item.title}</p>
                <p className="truncate text-xs text-text-secondary">
                  {item.artistName} · {formatSize(item.sizeBytes)}
                </p>
              </div>
            </Link>
            <button onClick={() => handleDelete(item.songId)} aria-label={`Remove ${item.title} download`}>
              <Trash2 size={18} className="text-text-secondary" />
            </button>
          </li>
        ))}
      </ul>

      {!loading && items.length === 0 && (
        <p className="mt-6 text-sm text-text-secondary">
          Belum ada lagu yang di-download. Buka lagu dan tap ikon download untuk
          menyimpannya untuk didengar offline.
        </p>
      )}
    </main>
  );
}
