const defaultGatewayPublicOrigin = "http://127.0.0.1:3000";

function bareHTTPOrigin(value: string): URL {
  const origin = new URL(value);
  if (
    !["http:", "https:"].includes(origin.protocol) ||
    origin.username ||
    origin.password ||
    origin.pathname !== "/" ||
    origin.search ||
    origin.hash
  ) {
    throw new Error("GATEWAY_PUBLIC_ORIGIN must be an HTTP(S) origin");
  }
  return origin;
}

export function socialSignInURL(
  requestURL: string,
  gatewayPublicOrigin = defaultGatewayPublicOrigin,
): URL {
  const request = new URL(requestURL);
  const signIn = new URL("/sign-in", bareHTTPOrigin(gatewayPublicOrigin));
  signIn.searchParams.set(
    "redirect_url",
    `${request.pathname}${request.search}`,
  );
  return signIn;
}
