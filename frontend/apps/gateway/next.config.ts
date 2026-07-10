import { localServiceOrigins } from "@coach-connect/config";
import type { NextConfig } from "next";

const socialOrigin = process.env.SOCIAL_ORIGIN ?? localServiceOrigins.social;
const marketplaceOrigin =
  process.env.MARKETPLACE_ORIGIN ?? localServiceOrigins.marketplace;

const nextConfig: NextConfig = {
  allowedDevOrigins: ["127.0.0.1"],
  transpilePackages: [
    "@coach-connect/auth",
    "@coach-connect/config",

    "@coach-connect/go-api-client",
    "@coach-connect/trpc-client",
    "@coach-connect/trpc-contract",
    "@coach-connect/ui",
  ],
  async rewrites() {
    return [
      { source: "/feed", destination: `${socialOrigin}/feed` },
      { source: "/feed/:path*", destination: `${socialOrigin}/feed/:path*` },
      {
        source: "/people/:path*",
        destination: `${socialOrigin}/people/:path*`,
      },
      {
        source: "/profile/:path*",
        destination: `${socialOrigin}/profile/:path*`,
      },
      {
        source: "/messages/:path*",
        destination: `${socialOrigin}/messages/:path*`,
      },
      {
        source: "/notifications/:path*",
        destination: `${socialOrigin}/notifications/:path*`,
      },
      {
        source: "/social-static/:path*",
        destination: `${socialOrigin}/social-static/:path*`,
      },
      {
        source: "/marketplace",
        destination: `${marketplaceOrigin}/marketplace`,
      },
      {
        source: "/marketplace/:path*",
        destination: `${marketplaceOrigin}/marketplace/:path*`,
      },
      {
        source: "/seller/:path*",
        destination: `${marketplaceOrigin}/seller/:path*`,
      },
      {
        source: "/orders/:path*",
        destination: `${marketplaceOrigin}/orders/:path*`,
      },
      {
        source: "/marketplace-static/:path*",
        destination: `${marketplaceOrigin}/marketplace-static/:path*`,
      },
    ];
  },
};

export default nextConfig;
