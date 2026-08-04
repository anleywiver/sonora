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
    </main>
  );
}
