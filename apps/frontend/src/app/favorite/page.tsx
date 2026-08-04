"use client";

import { Heart } from "lucide-react";
import Link from "next/link";
import { useEffect, useState } from "react";

import { apiFetch } from "@/lib/api";

interface Favorite {
  id: string;
  type: "song" | "album" | "artist" | "playlist";
  target_id: string;
}

interface ResolvedFavorite extends Favorite {
  title: string;
  subtitle: string;
  // Only song/playlist have a detail page today — artist/album favorites
  // (not yet favoritable anywhere in the UI, only reachable via direct API
  // calls) resolve fine but render without a link rather than a 404.
  href: string | null;
}

async function resolveFavorite(f: Favorite): Promise<ResolvedFavorite | null> {
  try {
    switch (f.type) {
      case "song": {
        const song = await apiFetch<{ title: string; artist_name: string }>(`/songs/${f.target_id}`);
        return { ...f, title: song.title, subtitle: song.artist_name, href: `/song/${f.target_id}` };
      }
      case "playlist": {
        const playlist = await apiFetch<{ name: string; description: string }>(
          `/playlists/${f.target_id}`,
        );
        return {
          ...f,
          title: playlist.name,
          subtitle: playlist.description || "Playlist",
          href: `/library/${f.target_id}`,
        };
      }
      case "artist": {
        const artist = await apiFetch<{ name: string }>(`/artists/${f.target_id}`);
        return { ...f, title: artist.name, subtitle: "Artist", href: null };
      }
      case "album": {
        const album = await apiFetch<{ title: string; artist_name: string }>(
          `/albums/${f.target_id}`,
        );
        return { ...f, title: album.title, subtitle: album.artist_name, href: null };
      }
    }
  } catch {
    return null;
  }
}

export default function FavoritePage() {
  const [favorites, setFavorites] = useState<ResolvedFavorite[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<Favorite[]>("/favorites")
      .then(async (raw) => {
        const resolved = await Promise.all(raw.map(resolveFavorite));
        setFavorites(resolved.filter((f): f is ResolvedFavorite => f !== null));
      })
      .finally(() => setLoading(false));
  }, []);

  const handleUnfavorite = async (f: Favorite) => {
    await apiFetch("/favorites", {
      method: "DELETE",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ type: f.type, id: f.target_id }),
    });
    setFavorites((prev) => prev.filter((x) => x.id !== f.id));
  };

  return (
    <main className="min-h-screen px-4 pb-32 pt-6">
      <h1 className="text-xl font-bold">Favorite</h1>

      {loading && <p className="mt-6 text-sm text-text-secondary">Memuat...</p>}

      <ul className="mt-4 space-y-2">
        {favorites.map((f) => (
          <li
            key={f.id}
            className="flex items-center gap-3 rounded-[18px] border border-border bg-white/5 p-3.5 backdrop-blur-md"
          >
            {f.href ? (
              <Link href={f.href} className="flex min-w-0 flex-1 items-center gap-3">
                <div className="h-12 w-12 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{f.title}</p>
                  <p className="truncate text-xs text-text-secondary">{f.subtitle}</p>
                </div>
              </Link>
            ) : (
              <div className="flex min-w-0 flex-1 items-center gap-3">
                <div className="h-12 w-12 flex-shrink-0 rounded-xl bg-gradient-to-br from-accent to-primary" />
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{f.title}</p>
                  <p className="truncate text-xs text-text-secondary">{f.subtitle}</p>
                </div>
              </div>
            )}
            <button
              onClick={() => handleUnfavorite(f)}
              aria-label={`Unfavorite ${f.title}`}
              className="text-error"
            >
              <Heart size={18} fill="currentColor" />
            </button>
          </li>
        ))}
      </ul>

      {!loading && favorites.length === 0 && (
        <p className="mt-6 text-sm text-text-secondary">Belum ada favorite.</p>
      )}
    </main>
  );
}
