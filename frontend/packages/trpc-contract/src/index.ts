import type { GoApiClient } from "@coach-connect/go-api-client";
import { initTRPC } from "@trpc/server";
import superjson from "superjson";

export interface TRPCContext {
  api: GoApiClient;
}

const t = initTRPC.context<TRPCContext>().create({
  transformer: superjson,
});

const systemRouter = t.router({
  health: t.procedure.query(({ ctx }) => ctx.api.health()),
});

export const appRouter = t.router({
  system: systemRouter,
});

export type AppRouter = typeof appRouter;
