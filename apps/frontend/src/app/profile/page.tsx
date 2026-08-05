"use client";

import { LogOut } from "lucide-react";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";

import { apiFetch, ApiError } from "@/lib/api";
import { useAuthStore } from "@/store/auth";

interface Me {
  id: string;
  name: string;
  email: string;
  avatar_url: string;
  role: "owner" | "member";
}

// Screens-spec #22 (user-facing) — avatar+upload, editable name, read-only
// email, role badge, sign out. Avatar upload is a small client-resized
// data: URL, not a Drive upload — see ADR 0009 for why.
export default function ProfilePage() {
  const [me, setMe] = useState<Me | null>(null);
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const router = useRouter();
  const setSession = useAuthStore((s) => s.setSession);

  useEffect(() => {
    apiFetch<Me>("/auth/me")
      .then((data) => {
        setMe(data);
        setName(data.name);
      })
      .catch(() => setError("Gagal memuat profil."));
  }, []);

  function saveName(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setSaved(false);
    apiFetch<Me>("/auth/me", {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name }),
    })
      .then((data) => {
        setMe(data);
        setSaved(true);
        setTimeout(() => setSaved(false), 2000);
      })
      .catch((e) => setError(e instanceof ApiError ? e.message : "Gagal menyimpan nama."))
      .finally(() => setSaving(false));
  }

  function handlePhotoChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setError(null);

    const img = new Image();
    const reader = new FileReader();
    reader.onload = () => {
      img.onload = () => {
        // Resize to a small thumbnail client-side so the data URL stays
        // well under the backend's size cap (ADR 0009) — no server-side
        // image processing needed for something this small.
        const size = 128;
        const canvas = document.createElement("canvas");
        canvas.width = size;
        canvas.height = size;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          setUploading(false);
          setError("Browser tidak mendukung pemrosesan gambar.");
          return;
        }
        const scale = Math.max(size / img.width, size / img.height);
        const w = img.width * scale;
        const h = img.height * scale;
        ctx.drawImage(img, (size - w) / 2, (size - h) / 2, w, h);
        const dataUrl = canvas.toDataURL("image/jpeg", 0.85);

        apiFetch<Me>("/auth/me", {
          method: "PUT",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ avatar_url: dataUrl }),
        })
          .then(setMe)
          .catch((e) => setError(e instanceof ApiError ? e.message : "Gagal upload foto."))
          .finally(() => setUploading(false));
      };
      img.src = reader.result as string;
    };
    reader.readAsDataURL(file);
  }

  async function signOut() {
    await apiFetch("/auth/logout", { method: "POST" }).catch(() => {});
    setSession(null, null);
    router.replace("/login");
  }

  if (!me) {
    return (
      <main className="flex min-h-screen items-center justify-center">
        <p className="text-sm text-text-secondary">{error ?? "Memuat..."}</p>
      </main>
    );
  }

  return (
    <main className="min-h-screen px-6 pb-32 pt-10">
      <div className="flex flex-col items-center">
        <button
          onClick={() => fileInputRef.current?.click()}
          disabled={uploading}
          className="relative h-24 w-24 overflow-hidden rounded-full bg-gradient-to-br from-accent to-primary disabled:opacity-50"
          aria-label="Ganti foto profil"
        >
          {me.avatar_url ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={me.avatar_url} alt="" className="h-full w-full object-cover" />
          ) : null}
        </button>
        <input ref={fileInputRef} type="file" accept="image/*" onChange={handlePhotoChange} className="hidden" />
        <button
          onClick={() => fileInputRef.current?.click()}
          disabled={uploading}
          className="mt-2 text-xs text-accent disabled:opacity-50"
        >
          {uploading ? "Mengunggah…" : "Ganti foto"}
        </button>

        <span className="mt-4 rounded-full bg-accent/15 px-3 py-1 text-[11px] font-semibold capitalize text-accent">
          {me.role}
        </span>
      </div>

      <form onSubmit={saveName} className="mx-auto mt-8 max-w-sm space-y-3">
        <div>
          <label className="text-xs text-text-secondary">Nama</label>
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="mt-1 w-full rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm"
          />
        </div>
        <div>
          <label className="text-xs text-text-secondary">Email</label>
          <input
            disabled
            value={me.email}
            className="mt-1 w-full cursor-not-allowed rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm text-text-secondary"
          />
        </div>

        {error && <p className="text-sm text-error">{error}</p>}
        {saved && <p className="text-sm text-success">Tersimpan.</p>}

        <button
          type="submit"
          disabled={saving}
          className="w-full rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white disabled:opacity-50"
        >
          {saving ? "Menyimpan…" : "Simpan"}
        </button>
      </form>

      <button
        onClick={signOut}
        className="mx-auto mt-8 flex items-center gap-2 text-sm text-error"
      >
        <LogOut size={16} />
        Sign out
      </button>
    </main>
  );
}
