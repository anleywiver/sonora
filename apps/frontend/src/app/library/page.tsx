"use client";

import { Plus } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";

interface Playlist {
  id: string;
  name: string;
  description: string;
}

export default function LibraryPage() {
  const [playlists, setPlaylists] = useState<Playlist[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [newName, setNewName] = useState("");

  const load = () => {
    setLoading(true);
    apiFetch<Playlist[]>("/playlists")
      .then(setPlaylists)
      .catch(() => setError("Gagal memuat playlist."))
      .finally(() => setLoading(false));
  };

  useEffect(load, []);

  const handleCreate = async () => {
    if (!newName.trim()) return;
    try {
      await apiFetch<Playlist>("/playlists", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: newName.trim() }),
      });
      setNewName("");
      setCreating(false);
      load();
    } catch {
      setError("Gagal membuat playlist.");
    }
  };

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold">Library</h1>
        <button
          onClick={() => setCreating((v) => !v)}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-primary text-white"
          aria-label="New playlist"
        >
          <Plus size={18} />
        </button>
      </div>

      {creating && (
        <div className="mt-4 flex gap-2">
          <input
            autoFocus
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleCreate()}
            placeholder="Nama playlist"
            className="flex-1 rounded-2xl border border-border bg-white/5 px-4 py-2 text-sm outline-none placeholder:text-text-secondary"
          />
          <button
            onClick={handleCreate}
            className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-white"
          >
            Buat
          </button>
        </div>
      )}

      {loading && <p className="mt-6 text-sm text-text-secondary">Memuat...</p>}
      {error && <p className="mt-6 text-sm text-error">{error}</p>}

      <ul className="mt-4 space-y-2">
        {playlists.map((p) => (
          <li key={p.id}>
            <Link
              href={`/library/${p.id}`}
              className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
            >
              <div className="h-12 w-12 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{p.name}</p>
                {p.description && (
                  <p className="truncate text-xs text-text-secondary">{p.description}</p>
                )}
              </div>
            </Link>
          </li>
        ))}
      </ul>

      {!loading && playlists.length === 0 && (
        <p className="mt-6 text-sm text-text-secondary">Belum ada playlist.</p>
      )}
    </main>
  );
}
