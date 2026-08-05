"use client";

import { Pencil, Search, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { formatDuration } from "@/lib/utils";

interface AdminSong {
  id: string;
  title: string;
  artist_name: string;
  album_title: string;
  duration_ms: number;
  storage_provider: string;
  created_at: string;
}

// Screens-spec #23 (Sprint 14 sisipan, ADR 0010). Soft delete only — see
// the ADR for the deliberate scope limit (an existing playlist/favorite
// link to a deleted song isn't blocked from playing).
export default function SongsPage() {
  const [songs, setSongs] = useState<AdminSong[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [editing, setEditing] = useState<AdminSong | null>(null);
  const [confirmingDelete, setConfirmingDelete] = useState<AdminSong | null>(null);
  const [busyIds, setBusyIds] = useState<Set<string>>(new Set());

  function load(query: string) {
    setLoading(true);
    const qs = query ? `?search=${encodeURIComponent(query)}` : "";
    apiFetch<AdminSong[]>(`/admin/songs${qs}`)
      .then(setSongs)
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to load songs"))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    const handle = setTimeout(() => load(search), 300);
    return () => clearTimeout(handle);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [search]);

  function deleteSong(song: AdminSong) {
    setBusyIds((s) => new Set(s).add(song.id));
    apiFetch(`/admin/songs/${song.id}`, { method: "DELETE" })
      .then(() => {
        setSongs((prev) => prev.filter((s) => s.id !== song.id));
        setConfirmingDelete(null);
      })
      .catch((e) => setError(e instanceof ApiError ? e.message : "Failed to delete song"))
      .finally(() =>
        setBusyIds((s) => {
          const next = new Set(s);
          next.delete(song.id);
          return next;
        }),
      );
  }

  return (
    <main className="min-h-screen px-8 py-8">
      <h1 className="text-2xl font-bold">Songs</h1>
      <p className="mt-1 text-sm text-text-secondary">Kelola katalog lagu.</p>

      <div className="mt-4 flex items-center gap-2 rounded-control border border-border bg-card px-4 py-2.5">
        <Search size={16} className="text-text-secondary" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Cari judul atau artist"
          className="w-full bg-transparent text-sm outline-none placeholder:text-text-secondary"
        />
      </div>

      {error && (
        <p className="mt-4 rounded-control border border-error/30 bg-error/10 px-4 py-2 text-sm text-error">
          {error}
        </p>
      )}

      <div className="mt-6 overflow-x-auto rounded-card border border-border bg-card">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left text-xs text-text-secondary">
              <th className="px-4 py-3 font-medium">Judul</th>
              <th className="px-4 py-3 font-medium">Artist</th>
              <th className="px-4 py-3 font-medium">Album</th>
              <th className="px-4 py-3 font-medium">Durasi</th>
              <th className="px-4 py-3 font-medium">Storage</th>
              <th className="px-4 py-3 font-medium">Ditambahkan</th>
              <th className="px-4 py-3 font-medium">Action</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={7} className="px-4 py-6 text-center text-text-secondary">
                  Loading…
                </td>
              </tr>
            ) : songs.length === 0 ? (
              <tr>
                <td colSpan={7} className="px-4 py-6 text-center text-text-secondary">
                  Tidak ada lagu.
                </td>
              </tr>
            ) : (
              songs.map((s) => (
                <tr key={s.id} className="border-b border-border last:border-0">
                  <td className="px-4 py-3">{s.title}</td>
                  <td className="px-4 py-3 text-text-secondary">{s.artist_name}</td>
                  <td className="px-4 py-3 text-text-secondary">{s.album_title || "—"}</td>
                  <td className="px-4 py-3 text-text-secondary">{formatDuration(s.duration_ms)}</td>
                  <td className="px-4 py-3 text-text-secondary capitalize">{s.storage_provider}</td>
                  <td className="px-4 py-3 text-text-secondary">{s.created_at}</td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button
                        onClick={() => setEditing(s)}
                        aria-label={`Edit ${s.title}`}
                        className="flex items-center justify-center rounded-control border border-border px-3 py-1.5 text-text-secondary hover:text-text-primary"
                      >
                        <Pencil size={13} />
                      </button>
                      <button
                        onClick={() => setConfirmingDelete(s)}
                        disabled={busyIds.has(s.id)}
                        aria-label={`Delete ${s.title}`}
                        className="flex items-center justify-center rounded-control border border-border px-3 py-1.5 text-error hover:bg-error/10 disabled:opacity-50"
                      >
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {editing && (
        <EditModal
          song={editing}
          onClose={() => setEditing(null)}
          onSaved={() => {
            setEditing(null);
            load(search);
          }}
          onError={setError}
        />
      )}

      {confirmingDelete && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
          <div className="w-full max-w-sm rounded-card border border-border bg-card p-5">
            <h2 className="text-sm font-semibold">Hapus lagu?</h2>
            <p className="mt-1 text-xs text-text-secondary">
              &quot;{confirmingDelete.title}&quot; akan dihapus (soft delete) dan hilang dari pencarian. Tindakan ini
              tidak bisa dibatalkan dari UI.
            </p>
            <div className="mt-4 flex gap-2">
              <button
                onClick={() => setConfirmingDelete(null)}
                className="flex-1 rounded-control border border-border px-4 py-2 text-sm font-medium text-text-secondary"
              >
                Batal
              </button>
              <button
                onClick={() => deleteSong(confirmingDelete)}
                disabled={busyIds.has(confirmingDelete.id)}
                className="flex-1 rounded-control bg-error px-4 py-2 text-sm font-semibold text-white disabled:opacity-50"
              >
                Hapus
              </button>
            </div>
          </div>
        </div>
      )}
    </main>
  );
}

function EditModal({
  song,
  onClose,
  onSaved,
  onError,
}: {
  song: AdminSong;
  onClose: () => void;
  onSaved: () => void;
  onError: (msg: string) => void;
}) {
  const [title, setTitle] = useState(song.title);
  const [artistName, setArtistName] = useState(song.artist_name);
  const [albumTitle, setAlbumTitle] = useState(song.album_title);
  const [genreName, setGenreName] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function submit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    apiFetch(`/admin/songs/${song.id}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        title,
        artist_name: artistName,
        album_title: albumTitle || null,
        genre_name: genreName || null,
      }),
    })
      .then(onSaved)
      .catch((e) => onError(e instanceof ApiError ? e.message : "Failed to update song"))
      .finally(() => setSubmitting(false));
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 px-4">
      <form onSubmit={submit} className="w-full max-w-sm rounded-card border border-border bg-card p-5">
        <h2 className="text-sm font-semibold">Edit metadata</h2>
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Judul"
          className="mt-4 w-full rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <input
          value={artistName}
          onChange={(e) => setArtistName(e.target.value)}
          placeholder="Artist"
          className="mt-3 w-full rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <input
          value={albumTitle}
          onChange={(e) => setAlbumTitle(e.target.value)}
          placeholder="Album (opsional)"
          className="mt-3 w-full rounded-control border border-border bg-background px-3 py-2 text-sm"
        />
        <input
          value={genreName}
          onChange={(e) => setGenreName(e.target.value)}
          placeholder="Genre (opsional, kosongkan = tidak diubah)"
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
            {submitting ? "Menyimpan…" : "Simpan"}
          </button>
        </div>
      </form>
    </div>
  );
}
