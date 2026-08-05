import type { LucideIcon } from "lucide-react";

// ui-implementation-spec.md #2.4 — shared empty-list treatment (icon
// badge + message + optional CTA), used wherever a list can be genuinely
// empty rather than each page inventing its own ad-hoc "belum ada..." text.
export function EmptyState({
  icon: Icon,
  message,
  ctaLabel,
  onCtaClick,
}: {
  icon: LucideIcon;
  message: string;
  ctaLabel?: string;
  onCtaClick?: () => void;
}) {
  return (
    <div className="flex min-h-[400px] flex-1 flex-col items-center justify-center px-8 text-center">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl border border-border bg-white/5">
        <Icon className="h-7 w-7 text-text-secondary" />
      </div>
      <p className="mb-5 max-w-[240px] text-sm text-text-secondary">{message}</p>
      {ctaLabel && (
        <button
          onClick={onCtaClick}
          className="rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white transition-colors hover:bg-hover"
        >
          {ctaLabel}
        </button>
      )}
    </div>
  );
}
