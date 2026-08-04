import Link from "next/link";

export default function HomePage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center px-6 pb-32 text-center">
      <div className="mx-auto mb-4 h-16 w-16 rounded-2xl bg-gradient-to-br from-accent to-primary" />
      <h1 className="text-2xl font-bold">Sonora</h1>
      <p className="mt-2 text-sm text-text-secondary">
        Home (Continue Listening, Trending, dst.) dibangun di Sprint 5-6 setelah
        data playlist/favorite/history ada.
      </p>
      <Link
        href="/search"
        className="mt-6 rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white"
      >
        Cari lagu
      </Link>
    </main>
  );
}
