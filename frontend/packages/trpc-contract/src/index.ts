import {
  canAccess,
  type AccessContext,
  type AccessRequirement,
  type AccessTier,
} from "@coach-connect/auth/access";
import type { GoApiClient } from "@coach-connect/go-api-client";
import { initTRPC, TRPCError } from "@trpc/server";
import superjson from "superjson";

export interface TRPCContext {
  access: AccessContext;
  api: GoApiClient;
}

const t = initTRPC.context<TRPCContext>().create({
  transformer: superjson,
});

export const router = t.router;
export const publicProcedure = t.procedure;

function accessProcedure(requirement: AccessRequirement) {
  return publicProcedure.use(({ ctx, next }) => {
    const access = ctx.access;

    if (!access.isAuthenticated) {
      throw new TRPCError({ code: "UNAUTHORIZED" });
    }

    if (!canAccess(access, requirement)) {
      throw new TRPCError({ code: "FORBIDDEN" });
    }

    return next({ ctx: { ...ctx, access } });
  });
}

export const authenticatedProcedure = accessProcedure({
  kind: "authenticated",
});

export const adminProcedure = accessProcedure({ kind: "admin-only" });

export interface MinimumTierProcedureOptions {
  allowAdminOverride?: boolean;
}

export function minimumTierProcedure(
  tier: AccessTier,
  options: MinimumTierProcedureOptions = {},
) {
  return accessProcedure({
    kind: "minimum-tier",
    tier,
    ...(options.allowAdminOverride === undefined
      ? {}
      : { allowAdminOverride: options.allowAdminOverride }),
  });
}

const systemRouter = router({
  health: publicProcedure.query(({ ctx }) => ctx.api.health()),
});

export const appRouter = router({
  system: systemRouter,
});

export type AppRouter = typeof appRouter;
