import {
  canAccess,
  type AccessContext,
  type AccessRequirement,
  type AccessTier,
} from "@coach-connect/auth/access";
import { GoApiError, type GoApiClient } from "@coach-connect/go-api-client";
import { initTRPC, TRPCError } from "@trpc/server";
import superjson from "superjson";
import { z } from "zod";

export interface TRPCContext {
  access: AccessContext;
  api: GoApiClient;
  bearerToken: string | null;
  identity: { displayName: string; avatarUrl: string | null } | null;
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

function auth(ctx: TRPCContext) {
  if (!ctx.access.isAuthenticated || !ctx.bearerToken)
    throw new TRPCError({ code: "UNAUTHORIZED" });
  return { bearerToken: ctx.bearerToken };
}

function mapGoError(error: unknown): never {
  if (error instanceof TRPCError) throw error;
  if (error instanceof GoApiError) {
    const code =
      error.code === "unauthenticated"
        ? "UNAUTHORIZED"
        : error.code === "forbidden"
          ? "FORBIDDEN"
          : error.code === "not_found"
            ? "NOT_FOUND"
            : error.code === "invalid_request"
              ? "BAD_REQUEST"
              : "INTERNAL_SERVER_ERROR";
    throw new TRPCError({ code, cause: error });
  }
  throw new TRPCError({ code: "INTERNAL_SERVER_ERROR", cause: error });
}

const pageInput = z
  .object({
    cursor: z.string().max(512).optional(),
    limit: z.number().int().min(1).max(50).default(20),
  })
  .default({ limit: 20 });
const id = z.string().uuid();
const postBody = z.string().trim().min(1).max(2000);
const commentBody = z.string().trim().min(1).max(1000);
const protectedCall = <T>(call: () => Promise<T>) => call().catch(mapGoError);

const socialRouter = router({
  feed: authenticatedProcedure
    .input(pageInput)
    .query(({ ctx, input }) =>
      protectedCall(() => ctx.api.listFeed({ ...auth(ctx), ...input })),
    ),
  saved: authenticatedProcedure
    .input(pageInput)
    .query(({ ctx, input }) =>
      protectedCall(() => ctx.api.listSavedPosts({ ...auth(ctx), ...input })),
    ),
  post: authenticatedProcedure
    .input(z.object({ postId: id }))
    .query(({ ctx, input }) =>
      protectedCall(() => ctx.api.getPost(input.postId, auth(ctx))),
    ),
  createPost: authenticatedProcedure
    .input(z.object({ body: postBody }))
    .mutation(({ ctx, input }) =>
      protectedCall(() => ctx.api.createPost(input.body, auth(ctx))),
    ),
  updatePost: authenticatedProcedure
    .input(z.object({ postId: id, body: postBody }))
    .mutation(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.updatePost(input.postId, input.body, auth(ctx)),
      ),
    ),
  deletePost: authenticatedProcedure
    .input(z.object({ postId: id }))
    .mutation(({ ctx, input }) =>
      protectedCall(() => ctx.api.deletePost(input.postId, auth(ctx))),
    ),
  setLike: authenticatedProcedure
    .input(z.object({ postId: id, liked: z.boolean() }))
    .mutation(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.setPostLike(input.postId, input.liked, auth(ctx)),
      ),
    ),
  setSaved: authenticatedProcedure
    .input(z.object({ postId: id, saved: z.boolean() }))
    .mutation(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.setSavedPost(input.postId, input.saved, auth(ctx)),
      ),
    ),
  comments: authenticatedProcedure
    .input(z.object({ postId: id }).and(pageInput))
    .query(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.listComments(input.postId, {
          ...auth(ctx),
          cursor: input.cursor,
          limit: input.limit,
        }),
      ),
    ),
  createComment: authenticatedProcedure
    .input(z.object({ postId: id, body: commentBody }))
    .mutation(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.createComment(input.postId, input.body, auth(ctx)),
      ),
    ),
  updateComment: authenticatedProcedure
    .input(z.object({ commentId: id, body: commentBody }))
    .mutation(({ ctx, input }) =>
      protectedCall(() =>
        ctx.api.updateComment(input.commentId, input.body, auth(ctx)),
      ),
    ),
  deleteComment: authenticatedProcedure
    .input(z.object({ commentId: id }))
    .mutation(({ ctx, input }) =>
      protectedCall(() => ctx.api.deleteComment(input.commentId, auth(ctx))),
    ),
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
  social: socialRouter,
});

export type AppRouter = typeof appRouter;
