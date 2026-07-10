import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  allowedDevOrigins: ["127.0.0.1"],
  assetPrefix: "/marketplace-static",
  transpilePackages: ["@coach-connect/auth", "@coach-connect/ui"],
};

export default nextConfig;
