import type { NextConfig } from "next";

/**
 * API proxying is handled by src/app/api/[...path]/route.ts when
 * API_PROXY_TARGET is set (see env/prod.env). That keeps /api same-origin
 * for first-party cookies and surfaces clearer errors than config rewrites.
 */
const nextConfig: NextConfig = {
  reactStrictMode: true,
};

export default nextConfig;
