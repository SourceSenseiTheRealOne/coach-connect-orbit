import type { GoApiClient } from "@coach-connect/go-api-client";
import { describe, expect, it, vi } from "vitest";
import { appRouter } from "./index";

describe("health procedure", () => {
  it("delegates liveness to the Go API client", async () => {
    const api: GoApiClient = {
      health: vi.fn().mockResolvedValue({
        status: "ok",
        service: "coach-connect-api",
      }),
    };
    const caller = appRouter.createCaller({ api });

    const result = await caller.system.health();

    expect(result).toEqual({ status: "ok", service: "coach-connect-api" });
    expect(api.health).toHaveBeenCalledOnce();
  });
});
