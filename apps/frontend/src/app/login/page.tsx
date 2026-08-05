const API_BASE = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080/api/v1";

export default function LoginPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-8 px-6">
      <div className="text-center">
        <div className="mx-auto mb-4 h-16 w-16 rounded-2xl bg-gradient-to-br from-accent to-primary" />
        <h1 className="text-2xl font-bold">Sonora</h1>
        <p className="mt-2 text-sm text-text-secondary">Personal music streaming</p>
      </div>

      <a
        href={`${API_BASE}/auth/google`}
        className="flex w-full max-w-xs items-center justify-center gap-3 rounded-2xl bg-white px-6 py-3 font-semibold text-black"
      >
        Continue with Google
      </a>

      {/* Apple sign-in has no backend (no Apple OAuth client configured) —
          shown disabled per screens-spec #2 rather than omitted, since the
          spec explicitly calls for the placeholder to communicate "coming
          later", not "doesn't exist". */}
      <button
        disabled
        className="flex w-full max-w-xs cursor-not-allowed items-center justify-center gap-3 rounded-2xl border border-border bg-white/5 px-6 py-3 font-semibold text-text-secondary opacity-50"
      >
        Continue with Apple
      </button>

      <div className="flex w-full max-w-xs items-center gap-3 text-xs text-text-secondary">
        <div className="h-px flex-1 bg-border" />
        or continue with email
        <div className="h-px flex-1 bg-border" />
      </div>

      {/* No email/password endpoint exists (Sprint 4 decision — see
          CLAUDE.md) — shown disabled, same placeholder treatment as Apple
          above, rather than a divider that leads to nothing. */}
      <div className="w-full max-w-xs space-y-3 opacity-50">
        <input
          disabled
          placeholder="Email"
          className="w-full cursor-not-allowed rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm placeholder:text-text-secondary"
        />
        <input
          disabled
          placeholder="Password"
          className="w-full cursor-not-allowed rounded-2xl border border-border bg-white/5 px-4 py-3 text-sm placeholder:text-text-secondary"
        />
        <button
          disabled
          className="w-full cursor-not-allowed rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white"
        >
          Sign in
        </button>
      </div>
    </main>
  );
}
