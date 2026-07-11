import { describe, expect, it, vi } from "vitest";
import { GoApiError, createGoApiClient } from "./index";

describe("createGoApiClient", () => {
  it("loads health through the configured Go API base URL", async () => {
    const fetchImplementation = vi
      .fn<typeof fetch>()
      .mockResolvedValue(
        new Response(
          JSON.stringify({ status: "ok", service: "coach-connect-api" }),
          { status: 200, headers: { "content-type": "application/json" } },
        ),
      );
    const client = createGoApiClient({
      baseUrl: "http://localhost:9000/",
      fetchImplementation,
    });

    const result = await client.health();

    expect(result).toEqual({ status: "ok", service: "coach-connect-api" });
    expect(fetchImplementation).toHaveBeenCalledWith(
      "http://localhost:9000/api/v1/health",
      expect.objectContaining({
        headers: expect.objectContaining({ accept: "application/json" }),
        method: "GET",
      }),
    );
  });

  it("forwards bearer tokens only through semantic social methods", async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ items: [] }), {
        status: 200,
        headers: { "content-type": "application/json" },
      }),
    );
    const client = createGoApiClient({
      baseUrl: "http://localhost:9000/",
      fetchImplementation,
    });

    await client.listFeed({ bearerToken: "session-token", limit: 10 });

    expect(fetchImplementation).toHaveBeenCalledWith(
      "http://localhost:9000/api/v1/feed?limit=10",
      expect.objectContaining({
        headers: expect.objectContaining({
          authorization: "Bearer session-token",
        }),
        method: "GET",
      }),
    );
  });

  it("accepts empty successful delete responses", async () => {
    const fetchImplementation = vi
      .fn<typeof fetch>()
      .mockResolvedValue(new Response(null, { status: 200 }));
    const client = createGoApiClient({
      baseUrl: "http://localhost:9000",
      fetchImplementation,
    });

    await expect(
      client.deletePost("post-id", { bearerToken: "session-token" }),
    ).resolves.toBeUndefined();
    await expect(
      client.deleteComment("comment-id", { bearerToken: "session-token" }),
    ).resolves.toBeUndefined();

    expect(fetchImplementation).toHaveBeenNthCalledWith(
      1,
      "http://localhost:9000/api/v1/posts/post-id",
      expect.objectContaining({
        headers: expect.objectContaining({
          authorization: "Bearer session-token",
        }),
        method: "DELETE",
      }),
    );
    expect(fetchImplementation).toHaveBeenNthCalledWith(
      2,
      "http://localhost:9000/api/v1/comments/comment-id",
      expect.objectContaining({ method: "DELETE" }),
    );
  });

  it("maps stable Go error envelopes", async () => {
    const fetchImplementation = vi.fn<typeof fetch>().mockResolvedValue(
      new Response(JSON.stringify({ code: "forbidden" }), {
        status: 403,
        headers: { "content-type": "application/json" },
      }),
    );
    const client = createGoApiClient({
      baseUrl: "http://localhost:9000",
      fetchImplementation,
    });

    await expect(
      client.createPost("body", { bearerToken: "session-token" }),
    ).rejects.toEqual(new GoApiError(403, "forbidden"));
  });
});
