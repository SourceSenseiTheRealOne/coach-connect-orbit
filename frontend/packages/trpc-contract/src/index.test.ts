import type { GoApiClient } from "@coach-connect/go-api-client";
import { createAccessContext } from "@coach-connect/auth/access";
import { TRPCError } from "@trpc/server";
import { describe, expect, it, vi } from "vitest";
import {
  adminProcedure,
  appRouter,
  authenticatedProcedure,
  minimumTierProcedure,
  router,
  type TRPCContext,
} from "./index";

function createContext(
  access = createAccessContext({ userId: null }),
): TRPCContext {
  const api: GoApiClient = {
    health: vi.fn().mockResolvedValue({
      status: "ok",
      service: "coach-connect-api",
    }),
  };

  return { access, api };
}

describe("health procedure", () => {
  it("delegates liveness to the Go API client", async () => {
    const context = createContext();
    const caller = appRouter.createCaller(context);

    const result = await caller.system.health();

    expect(result).toEqual({ status: "ok", service: "coach-connect-api" });
    expect(context.api.health).toHaveBeenCalledOnce();
  });
});

describe("access procedures", () => {
  const guardedRouter = router({
    account: authenticatedProcedure.query(({ ctx }) => ctx.access.userId),
    moderation: adminProcedure.query(() => "moderation"),
    proFeature: minimumTierProcedure("pro").query(() => "pro"),
    proFeatureWithAdminOverride: minimumTierProcedure("pro", {
      allowAdminOverride: true,
    }).query(() => "pro-or-admin"),
  });

  it("returns UNAUTHORIZED for a guest", async () => {
    const error = await guardedRouter
      .createCaller(createContext())
      .account()
      .catch((caught: unknown) => caught);

    expect(error).toBeInstanceOf(TRPCError);
    expect(error).toMatchObject({ code: "UNAUTHORIZED" });
  });

  it("returns FORBIDDEN when an authenticated user lacks the policy", async () => {
    const caller = guardedRouter.createCaller(
      createContext(
        createAccessContext({ userId: "free", role: "user", tier: "free" }),
      ),
    );

    await expect(caller.proFeature()).rejects.toMatchObject({
      code: "FORBIDDEN",
    });
    await expect(caller.moderation()).rejects.toMatchObject({
      code: "FORBIDDEN",
    });
  });

  it("allows a paid tier to use its protected procedure", async () => {
    const caller = guardedRouter.createCaller(
      createContext(
        createAccessContext({ userId: "pro", role: "user", tier: "pro" }),
      ),
    );

    await expect(caller.proFeature()).resolves.toBe("pro");
  });

  it("requires explicit admin override for tier-protected procedures", async () => {
    const caller = guardedRouter.createCaller(
      createContext(
        createAccessContext({
          userId: "admin",
          role: "admin",
          tier: "free",
        }),
      ),
    );

    await expect(caller.proFeature()).rejects.toMatchObject({
      code: "FORBIDDEN",
    });
    await expect(caller.proFeatureWithAdminOverride()).resolves.toBe(
      "pro-or-admin",
    );
    await expect(caller.moderation()).resolves.toBe("moderation");
  });
});
