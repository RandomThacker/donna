import { type NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";
export const runtime = "nodejs";

const HOP_BY_HOP = new Set([
  "connection",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailers",
  "transfer-encoding",
  "upgrade",
  "host",
  "content-length",
]);

function getProxyTarget(): string | null {
  const raw = process.env.API_PROXY_TARGET?.trim();
  if (!raw) {
    return null;
  }
  return raw.replace(/\/$/, "");
}

async function proxyRequest(
  req: NextRequest,
  pathSegments: string[],
): Promise<NextResponse> {
  const target = getProxyTarget();
  if (!target) {
    return new NextResponse(
      "API_PROXY_TARGET is not set. Run `npm run env:prod` and restart Next.",
      { status: 500 },
    );
  }

  const incoming = new URL(req.url);
  const upstreamURL = `${target}/api/${pathSegments.join("/")}${incoming.search}`;

  const headers = new Headers();
  req.headers.forEach((value, key) => {
    if (!HOP_BY_HOP.has(key.toLowerCase())) {
      headers.set(key, value);
    }
  });

  const init: RequestInit = {
    method: req.method,
    headers,
    redirect: "manual",
  };

  if (req.method !== "GET" && req.method !== "HEAD") {
    init.body = req.body;
    // Required when streaming a request body in Node fetch.
    (init as RequestInit & { duplex?: string }).duplex = "half";
  }

  try {
    const upstream = await fetch(upstreamURL, init);
    const outHeaders = new Headers();
    upstream.headers.forEach((value, key) => {
      const lower = key.toLowerCase();
      if (HOP_BY_HOP.has(lower) || lower === "content-encoding") {
        return;
      }
      // Node fetch may combine Set-Cookie; append preserves multiple cookies.
      if (lower === "set-cookie") {
        outHeaders.append(key, value);
        return;
      }
      outHeaders.set(key, value);
    });

    return new NextResponse(upstream.body, {
      status: upstream.status,
      statusText: upstream.statusText,
      headers: outHeaders,
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : "unknown proxy error";
    console.error("[donna-api-proxy]", upstreamURL, message);
    return new NextResponse(
      `API proxy failed talking to ${target}.\n${message}\n\nIf curl to Railway works but this fails, corporate SSL/VPN is likely intercepting Node. Try: NODE_TLS_REJECT_UNAUTHORIZED=0 npm run dev`,
      { status: 502 },
    );
  }
}

type RouteContext = {
  params: Promise<{ path: string[] }>;
};

async function handle(req: NextRequest, context: RouteContext) {
  const { path } = await context.params;
  return proxyRequest(req, path ?? []);
}

export const GET = handle;
export const POST = handle;
export const PUT = handle;
export const PATCH = handle;
export const DELETE = handle;
export const OPTIONS = handle;
export const HEAD = handle;
