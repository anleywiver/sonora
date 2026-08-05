// Same brand mark as apps/frontend/src/components/logo.tsx — small,
// deliberate duplication (separate Next.js app, no shared UI package
// between frontend/admin yet) rather than a premature cross-app
// abstraction for one component.
export function Logo({ size = 64, className }: { size?: number; className?: string }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 512 512"
      className={className}
      role="img"
      aria-label="Sonora"
    >
      <defs>
        <linearGradient id="sonora-logo-g" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" stopColor="#3B82F6" />
          <stop offset="100%" stopColor="#1D4ED8" />
        </linearGradient>
      </defs>
      <rect x="56" y="56" width="400" height="400" rx="88" fill="url(#sonora-logo-g)" />
      <g fill="#FFFFFF">
        <rect x="88" y="186" width="48" height="140" rx="24" />
        <rect x="160" y="146" width="48" height="220" rx="24" />
        <rect x="232" y="96" width="48" height="320" rx="24" />
        <rect x="304" y="146" width="48" height="220" rx="24" />
        <rect x="376" y="186" width="48" height="140" rx="24" />
      </g>
    </svg>
  );
}
