import { describe, expect, it, vi } from "vitest";
import { createGoApiClient } from "./index";

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
});
