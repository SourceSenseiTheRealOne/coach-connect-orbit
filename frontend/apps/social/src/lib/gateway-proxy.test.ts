import { describe, expect, it, vi } from "vitest";
import { proxyGatewayRequest } from "./gateway-proxy";

describe("proxyGatewayRequest", () => {
  it("forwards tRPC path, query, auth, cookies, and body to the gateway", async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(
      new Response("proxied", {
        headers: {
          "content-type": "application/json",
          "x-middleware-rewrite": "/api/trpc/social.createPost",
        },
        status: 200,
      }),
    );
    const request = new Request(
      "http://social:3001/api/trpc/social.createPost?batch=1",
      {
        method: "POST",
        headers: {
          authorization: "Bearer session-token",
          connection: "keep-alive",
          cookie: "__session=clerk-session",
          "content-length": "17",
          "content-type": "application/json",
          host: "social:3001",
        },
        body: JSON.stringify({ body: "goal" }),
      },
    );

    const response = await proxyGatewayRequest(
      request,
      "http://gateway:3000",
      fetchImplementation,
    );

    expect(await response.text()).toBe("proxied");
    expect(response.headers.get("content-type")).toBe("application/json");
    expect(response.headers.get("x-middleware-rewrite")).toBeNull();
    expect(fetchImplementation).toHaveBeenCalledOnce();
    const [target, init] = fetchImplementation.mock.calls[0]!;
    expect(String(target)).toBe(
      "http://gateway:3000/api/trpc/social.createPost?batch=1",
    );
    const headers = new Headers(init?.headers);
    expect(headers.get("authorization")).toBe("Bearer session-token");
    expect(headers.get("cookie")).toBe("__session=clerk-session");
    expect(headers.get("host")).toBeNull();
    expect(headers.get("connection")).toBeNull();
    expect(headers.get("content-length")).toBeNull();
    expect(new TextDecoder().decode(init?.body as ArrayBuffer)).toBe(
      JSON.stringify({ body: "goal" }),
    );
  });

  it("rejects configured URLs that are not bare HTTP origins", async () => {
    await expect(
      proxyGatewayRequest(
        new Request("http://social:3001/api/trpc/system.health"),
        "file:///tmp/gateway",
      ),
    ).rejects.toThrow("GATEWAY_ORIGIN must be an HTTP(S) origin");
  });
});
