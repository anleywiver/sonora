"use client";

import { Heart } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { EmptyState } from "@/components/empty-state";
import { apiFetch } from "@/lib/api";
import { cn } from "@/lib/utils";

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

type Tab = "song" | "album" | "artist" | "playlist";
const TABS: { key: Tab; label: string }[] = [
  { key: "song", label: "Songs" },
  { key: "album", label: "Albums" },
  { key: "artist", label: "Artists" },
  { key: "playlist", label: "Playlists" },
];

export default function FavoritePage() {
  const router = useRouter();
  const [favorites, setFavorites] = useState<ResolvedFavorite[]>([]);
  const [loading, setLoading] = useState(true);
  const [tab, setTab] = useState<Tab>("song");

  useEffect(() => {
    apiFetch<Favorite[]>("/favorites")
      .then(async (raw) => {
        const resolved = await Promise.all(raw.map(resolveFavorite));
        setFavorites(resolved.filter((f): f is ResolvedFavorite => f !== null));
      })
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => favorites.filter((f) => f.type === tab), [favorites, tab]);

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

      <div className="mt-4 flex gap-2 overflow-x-auto">
        {TABS.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            className={cn(
              "flex-shrink-0 rounded-full px-4 py-1.5 text-xs font-medium",
              tab === t.key ? "bg-primary text-white" : "border border-border text-text-secondary",
            )}
          >
            {t.label}
          </button>
        ))}
      </div>

      {loading && <p className="mt-6 text-sm text-text-secondary">Memuat...</p>}

      <ul className="mt-4 space-y-2">
        {filtered.map((f) => (
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

      {!loading && filtered.length === 0 && (
        <EmptyState
          icon={Heart}
          message="Belum ada favorite. Mulai dengan menandai lagu yang kamu suka."
          ctaLabel="Cari lagu"
          onCtaClick={() => router.push("/search")}
        />
      )}
    </main>
  );
}
