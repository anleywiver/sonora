"use client";

import { usePathname } from "next/navigation";

import { Sidebar } from "@/components/sidebar";
import { useAuthStore } from "@/store/auth";

const NO_SHELL_PATHS = ["/login", "/auth/callback"];

export function AppShell({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const role = useAuthStore((s) => s.role);
  const showShell = !NO_SHELL_PATHS.includes(pathname);

  if (showShell && role !== null && role !== "owner") {
    return (
      <main className="flex min-h-screen items-center justify-center px-6">
        <div className="text-center">
          <h1 className="text-xl font-bold">Access Denied</h1>
          <p className="mt-2 text-sm text-text-secondary">
            Admin ini khusus untuk akun Owner.
          </p>
        </div>
      </main>
    );
  }

  return (
    <>
      {showShell && <Sidebar />}
      <div className={showShell ? "pl-60" : ""}>{children}</div>
    </>
  );
}
