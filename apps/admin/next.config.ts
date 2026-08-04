import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  reactStrictMode: true,
  transpilePackages: ["@sonora/ui", "@sonora/shared-types"],
};

export default nextConfig;
