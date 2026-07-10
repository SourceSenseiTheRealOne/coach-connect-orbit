import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  allowedDevOrigins: ["127.0.0.1"],
  assetPrefix: "/social-static",
  transpilePackages: ["@coach-connect/auth", "@coach-connect/ui"],
};

export default nextConfig;
