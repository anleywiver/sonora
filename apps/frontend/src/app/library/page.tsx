"use client";

import { Download, Plus, Search } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";
import { cn, formatDuration } from "@/lib/utils";
import { usePlayerStore } from "@/store/player";

interface Playlist {
  id: string;
  name: string;
  description: string;
}

interface LibrarySong {
  id: string;
  title: string;
  duration_ms: number;
  artist_name: string;
  album_title: string;
}

interface LibraryAlbum {
  id: string;
  title: string;
  cover_url: string;
  artist_name: string;
}

interface LibraryArtist {
  id: string;
  name: string;
  image_url: string;
}

type Tab = "songs" | "albums" | "artists" | "playlists";

// Screens-spec (Sprint 14 sisipan) — Browse Library shows the whole
// catalog per tab, not just favorited items (that's still /favorite).
export default function LibraryPage() {
  const [tab, setTab] = useState<Tab>("playlists");
  const [search, setSearch] = useState("");
  const [sortAlpha, setSortAlpha] = useState(false);

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <h1 className="text-xl font-bold">Library</h1>

      <div className="mt-4 flex gap-2 overflow-x-auto">
        {(["playlists", "songs", "albums", "artists"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={cn(
              "flex-shrink-0 rounded-full px-4 py-1.5 text-xs font-medium capitalize",
              tab === t ? "bg-primary text-white" : "border border-border text-text-secondary",
            )}
          >
            {t}
          </button>
        ))}
      </div>

      {tab !== "playlists" && (
        <div className="mt-4 flex items-center gap-2">
          <div className="flex flex-1 items-center gap-2 rounded-2xl border border-border bg-white/5 px-3 py-2">
            <Search size={14} className="text-text-secondary" />
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={`Cari ${tab}`}
              className="w-full bg-transparent text-sm outline-none placeholder:text-text-secondary"
            />
          </div>
          <button
            onClick={() => setSortAlpha((v) => !v)}
            className="rounded-2xl border border-border px-3 py-2 text-xs text-text-secondary"
          >
            {sortAlpha ? "A–Z" : "Terbaru"}
          </button>
        </div>
      )}

      <div className="mt-4">
        {tab === "playlists" && <PlaylistsTab />}
        {tab === "songs" && <SongsTab search={search} sortAlpha={sortAlpha} />}
        {tab === "albums" && <AlbumsTab search={search} sortAlpha={sortAlpha} />}
        {tab === "artists" && <ArtistsTab search={search} sortAlpha={sortAlpha} />}
      </div>
    </main>
  );
}

function PlaylistsTab() {
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
    <>
      <div className="flex items-center justify-between">
        <Link
          href="/downloads"
          className="flex flex-1 items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
        >
          <Download size={18} className="text-text-secondary" />
          <span className="text-sm font-medium">Downloads</span>
        </Link>
        <button
          onClick={() => setCreating((v) => !v)}
          className="ml-2 flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-full bg-primary text-white"
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
          <button onClick={handleCreate} className="rounded-2xl bg-primary px-4 py-2 text-sm font-semibold text-white">
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
                {p.description && <p className="truncate text-xs text-text-secondary">{p.description}</p>}
              </div>
            </Link>
          </li>
        ))}
      </ul>

      {!loading && playlists.length === 0 && <p className="mt-6 text-sm text-text-secondary">Belum ada playlist.</p>}
    </>
  );
}

function SongsTab({ search, sortAlpha }: { search: string; sortAlpha: boolean }) {
  const [songs, setSongs] = useState<LibrarySong[] | null>(null);
  const play = usePlayerStore((s) => s.play);

  useEffect(() => {
    const handle = setTimeout(() => {
      apiFetch<LibrarySong[]>(`/library/songs?search=${encodeURIComponent(search)}&sort=${sortAlpha ? "alpha" : "recent"}`)
        .then(setSongs)
        .catch(() => setSongs([]));
    }, 300);
    return () => clearTimeout(handle);
  }, [search, sortAlpha]);

  if (!songs) return <p className="text-sm text-text-secondary">Memuat...</p>;
  if (songs.length === 0) return <p className="text-sm text-text-secondary">Tidak ada lagu.</p>;

  return (
    <ul className="space-y-2">
      {songs.map((s) => (
        <li key={s.id}>
          <button
            onClick={() => void play({ id: s.id, title: s.title, artistName: s.artist_name, albumTitle: s.album_title, durationMs: s.duration_ms })}
            className="flex w-full items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 text-left backdrop-blur-md"
          >
            <div className="h-10 w-10 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
            <div className="min-w-0 flex-1">
              <p className="truncate text-sm font-medium">{s.title}</p>
              <p className="truncate text-xs text-text-secondary">{s.artist_name}</p>
            </div>
            <span className="text-xs text-text-secondary">{formatDuration(s.duration_ms)}</span>
          </button>
        </li>
      ))}
    </ul>
  );
}

function AlbumsTab({ search, sortAlpha }: { search: string; sortAlpha: boolean }) {
  const [albums, setAlbums] = useState<LibraryAlbum[] | null>(null);

  useEffect(() => {
    const handle = setTimeout(() => {
      apiFetch<LibraryAlbum[]>(`/library/albums?search=${encodeURIComponent(search)}&sort=${sortAlpha ? "alpha" : "recent"}`)
        .then(setAlbums)
        .catch(() => setAlbums([]));
    }, 300);
    return () => clearTimeout(handle);
  }, [search, sortAlpha]);

  if (!albums) return <p className="text-sm text-text-secondary">Memuat...</p>;
  if (albums.length === 0) return <p className="text-sm text-text-secondary">Tidak ada album.</p>;

  return (
    <div className="grid grid-cols-2 gap-3">
      {albums.map((a) => (
        <Link key={a.id} href={`/album/${a.id}`} className="rounded-[18px] border border-border bg-white/5 p-3 backdrop-blur-md">
          <div className="aspect-square w-full rounded-xl bg-gradient-to-br from-accent to-primary" />
          <p className="mt-2 truncate text-sm font-medium">{a.title}</p>
          <p className="truncate text-xs text-text-secondary">{a.artist_name}</p>
        </Link>
      ))}
    </div>
  );
}

function ArtistsTab({ search, sortAlpha }: { search: string; sortAlpha: boolean }) {
  const [artists, setArtists] = useState<LibraryArtist[] | null>(null);

  useEffect(() => {
    const handle = setTimeout(() => {
      apiFetch<LibraryArtist[]>(`/library/artists?search=${encodeURIComponent(search)}&sort=${sortAlpha ? "alpha" : "recent"}`)
        .then(setArtists)
        .catch(() => setArtists([]));
    }, 300);
    return () => clearTimeout(handle);
  }, [search, sortAlpha]);

  if (!artists) return <p className="text-sm text-text-secondary">Memuat...</p>;
  if (artists.length === 0) return <p className="text-sm text-text-secondary">Tidak ada artist.</p>;

  return (
    <ul className="space-y-2">
      {artists.map((a) => (
        <li key={a.id}>
          <Link href={`/artist/${a.id}`} className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md">
            <div className="h-10 w-10 flex-shrink-0 rounded-full bg-gradient-to-br from-accent to-primary" />
            <p className="truncate text-sm font-medium">{a.name}</p>
          </Link>
        </li>
      ))}
    </ul>
  );
}
