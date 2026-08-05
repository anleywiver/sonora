import { Logo } from "./logo";

// Screens-spec #1: logo on a gradient box, pulsing radial glow behind it,
// sequential bounce dots. Shown as the Providers loading state (session
// bootstrap) rather than a fixed artificial timer — it's on screen for as
// long as that actually takes, not a fake delay for spec's sake.
export function SplashScreen() {
  return (
    <main className="relative flex min-h-screen flex-col items-center justify-center overflow-hidden bg-background">
      <div
        className="absolute h-72 w-72 animate-pulse rounded-full bg-primary/30 blur-3xl"
        style={{ animationDuration: "2.5s" }}
      />
      <Logo size={80} className="relative" />
      <p className="relative mt-4 text-lg font-bold">Sonora</p>
      <div className="relative mt-6 flex gap-1.5">
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className="h-2 w-2 animate-bounce rounded-full bg-hover"
            style={{ animationDelay: `${i * 0.15}s` }}
          />
        ))}
      </div>
    </main>
  );
}
