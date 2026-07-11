const hopByHopHeaders = [
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
] as const;

const internalResponseHeaders = [
  ...hopByHopHeaders,
  "x-middleware-next",
  "x-middleware-rewrite",
  "x-nextjs-matched-path",
  "x-nextjs-rewrite",
] as const;

function gatewayTarget(requestUrl: string, gatewayOrigin: string): URL {
  const origin = new URL(gatewayOrigin);
  if (
    !["http:", "https:"].includes(origin.protocol) ||
    origin.username ||
    origin.password ||
    origin.pathname !== "/" ||
    origin.search ||
    origin.hash
  ) {
    throw new Error("GATEWAY_ORIGIN must be an HTTP(S) origin");
  }

  const target = new URL(requestUrl);
  target.protocol = origin.protocol;
  target.host = origin.host;
  return target;
}

export async function proxyGatewayRequest(
  request: Request,
  gatewayOrigin: string,
  fetchImplementation: typeof fetch = fetch,
): Promise<Response> {
  const headers = new Headers(request.headers);
  for (const name of hopByHopHeaders) headers.delete(name);

  const init: RequestInit = {
    headers,
    method: request.method,
    redirect: "manual",
  };
  if (request.method !== "GET" && request.method !== "HEAD") {
    init.body = await request.arrayBuffer();
  }

  const upstream = await fetchImplementation(
    gatewayTarget(request.url, gatewayOrigin),
    init,
  );
  const responseHeaders = new Headers(upstream.headers);
  for (const name of internalResponseHeaders) responseHeaders.delete(name);
  return new Response(upstream.body, {
    headers: responseHeaders,
    status: upstream.status,
    statusText: upstream.statusText,
  });
}
