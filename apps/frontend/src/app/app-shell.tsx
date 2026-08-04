"use client";

import { usePathname } from "next/navigation";

import { BottomNav } from "@/components/bottom-nav";
import { MiniPlayer } from "@/components/mini-player";

const NO_SHELL_PATHS = ["/login", "/auth/callback"];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const showShell = !NO_SHELL_PATHS.includes(pathname);

  return (
    <>
      {children}
      {showShell && (
        <>
          <MiniPlayer />
          <BottomNav />
        </>
      )}
    </>
  );
}
