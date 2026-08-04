"use client";

import { Home, Search, Library, Heart, Settings } from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";

const items = [
  { href: "/", icon: Home, label: "Home" },
  { href: "/search", icon: Search, label: "Search" },
  { href: "/library", icon: Library, label: "Library" },
  { href: "/favorite", icon: Heart, label: "Favorite" },
  { href: "/settings", icon: Settings, label: "Settings" },
];

export function BottomNav() {
  const pathname = usePathname();

  return (
    <nav className="fixed bottom-2 left-3 right-3 z-40 mx-auto flex max-w-md items-center justify-around rounded-[22px] border border-border bg-card/85 px-5 py-2.5 backdrop-blur-md">
      {items.map(({ href, icon: Icon, label }) => {
        const active = pathname === href;
        return (
          <Link
            key={href}
            href={href}
            className={cn(
              "flex flex-col items-center gap-1 text-[11px]",
              active ? "text-hover" : "text-text-secondary",
            )}
          >
            <Icon size={22} />
            {label}
          </Link>
        );
      })}
    </nav>
  );
}
