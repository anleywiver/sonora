import Link from "next/link";

export default function SettingsPage() {
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 px-6 pb-32 text-center">
      <Link href="/profile" className="rounded-2xl bg-primary px-6 py-3 text-sm font-semibold text-white">
        Profile
      </Link>
      <p className="text-sm text-text-secondary">Settings lain dibangun di sprint lanjutan.</p>
    </main>
  );
}
