import type { NextConfig } from "next";
import withSerwistInit from "@serwist/next";

/**
 * API proxying is handled by src/app/api/[...path]/route.ts when
 * API_PROXY_TARGET is set (see env/prod.env). That keeps /api same-origin
 * for first-party cookies and surfaces clearer errors than config rewrites.
 */
const withSerwist = withSerwistInit({
  swSrc: "src/sw.ts",
  swDest: "public/sw.js",
  disable: process.env.NODE_ENV === "development",
});

const nextConfig: NextConfig = {
  reactStrictMode: true,
  webpack: (config) => {
    config.module.rules.push({
      test: /\.mp3$/,
      type: "asset/resource",
    });
    return config;
  },
};

export default withSerwist(nextConfig);
