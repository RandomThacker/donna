import type { NextConfig } from "next";

/**
 * When API_PROXY_TARGET is set (Vercel production), browser calls stay same-origin
 * on /api/* and Next rewrites them to Railway. That makes donna_session a
 * first-party cookie on the Vercel host — required now that browsers block
 * cross-site cookies between *.vercel.app and *.up.railway.app.
 */
const apiProxyTarget = process.env.API_PROXY_TARGET?.trim().replace(/\/$/, "");

const nextConfig: NextConfig = {
  reactStrictMode: true,
  async rewrites() {
    if (!apiProxyTarget) {
      return [];
    }
    return [
      {
        source: "/api/:path*",
        destination: `${apiProxyTarget}/api/:path*`,
      },
    ];
  },
};

export default nextConfig;
