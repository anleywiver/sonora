"use client";

import { usePathname } from "next/navigation";

import { BottomNav } from "@/components/bottom-nav";
import { MiniPlayer } from "@/components/mini-player";

const NO_SHELL_PATHS = ["/login", "/auth/callback"];

// Full-screen player views (ui-implementation-spec.md #6.1) — Mini Player
// and BottomNav were rendering globally with no exception for these,
// stacking underneath the full-screen takeover instead of being replaced
// by it (each of these three already has its own chevron-down/back
// affordance, same as a modal would).
const FULLSCREEN_PLAYER_PATHS = ["/now-playing", "/lyrics", "/queue"];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const showShell = !NO_SHELL_PATHS.includes(pathname) && !FULLSCREEN_PLAYER_PATHS.includes(pathname);

  return (
    <>
      {/* App ini didesain mobile-first (docs/design-system.md) — di layar
          lebar (desktop/tablet), konten dikunci ke lebar mobile (max-w-md,
          sama seperti BottomNav/MiniPlayer di bawah) dan di-center, bukan
          diregangkan penuh selebar viewport. Di dalam kolom ini tetap
          responsive penuh mengikuti breakpoint Tailwind biasa. */}
      <div className="relative mx-auto min-h-screen w-full max-w-md sm:border-x sm:border-border">
        {children}
      </div>
      {showShell && (
        <>
          <MiniPlayer />
          <BottomNav />
        </>
      )}
    </>
  );
}
