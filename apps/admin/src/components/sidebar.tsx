"use client";

import { LayoutDashboard, HardDrive, Radio, Mic2, ListChecks, BarChart3, Users, Music, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";

const items = [
  { href: "/", icon: LayoutDashboard, label: "Dashboard" },
  { href: "/drive-manager", icon: HardDrive, label: "Drive Manager" },
  { href: "/ingest-sources", icon: Radio, label: "Ingest Sources" },
  { href: "/lyrics-source", icon: Mic2, label: "Lyrics Source" },
  { href: "/job-queue", icon: ListChecks, label: "Job Queue" },
  { href: "/analytics", icon: BarChart3, label: "Analytics" },
  { href: "/users", icon: Users, label: "Users" },
  { href: "/songs", icon: Music, label: "Songs" },
  { href: "/settings", icon: Settings, label: "Settings" },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <nav className="fixed inset-y-0 left-0 z-40 flex w-60 flex-col gap-1 border-r border-border bg-card/60 px-3 py-6 backdrop-blur-md">
      <div className="mb-6 px-3">
        <p className="text-lg font-bold">Sonora</p>
        <p className="text-xs text-text-secondary">Admin</p>
      </div>
      {items.map(({ href, icon: Icon, label }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex items-center gap-3 rounded-control px-3 py-2.5 text-sm",
              active
                ? "bg-primary/15 text-hover"
                : "text-text-secondary hover:bg-card hover:text-text-primary",
            )}
          >
            <Icon size={18} />
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
