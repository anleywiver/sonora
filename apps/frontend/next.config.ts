import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@sonora/ui", "@sonora/shared-types"],
  // Sprint 14: self-contained production server bundle for the Nginx +
  // Docker Compose deploy (infrastructure/docker/frontend.Dockerfile).
  output: "standalone",
};

export default nextConfig;
