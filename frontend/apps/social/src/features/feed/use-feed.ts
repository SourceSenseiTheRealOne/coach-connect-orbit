"use client";

import type { FeedPage, Post } from "@coach-connect/go-api-client";
import { trpc } from "@coach-connect/trpc-client";
import { cancelFeedQueriesBeforeSnapshot } from "./optimistic-cache";
import { applyOptimisticLike, applyOptimisticSave } from "./optimistic";

export type FeedView = "latest" | "saved";

export function useFeed(view: FeedView) {
  const utils = trpc.useUtils();
  const query = (
    view === "saved" ? trpc.social.saved : trpc.social.feed
  ).useInfiniteQuery(
    { limit: 10 },
    { getNextPageParam: (page) => page.nextCursor },
  );
  const invalidate = () =>
    Promise.all([
      utils.social.feed.invalidate(),
      utils.social.saved.invalidate(),
    ]);
  const create = trpc.social.createPost.useMutation({ onSuccess: invalidate });
  const update = trpc.social.updatePost.useMutation({ onSuccess: invalidate });
  const remove = trpc.social.deletePost.useMutation({ onSuccess: invalidate });

  async function optimistic(postId: string, transform: (post: Post) => Post) {
    await cancelFeedQueriesBeforeSnapshot(
      () => utils.social.feed.cancel(),
      () => utils.social.saved.cancel(),
    );
    const previousFeed = utils.social.feed.getInfiniteData({ limit: 10 });
    const previousSaved = utils.social.saved.getInfiniteData({ limit: 10 });
    const apply = (data: typeof previousFeed) =>
      data && {
        ...data,
        pages: data.pages.map((page: FeedPage) => ({
          ...page,
          items: page.items.map((post) =>
            post.id === postId ? transform(post) : post,
          ),
        })),
      };
    utils.social.feed.setInfiniteData({ limit: 10 }, apply);
    utils.social.saved.setInfiniteData({ limit: 10 }, apply);
    return () => {
      utils.social.feed.setInfiniteData({ limit: 10 }, previousFeed);
      utils.social.saved.setInfiniteData({ limit: 10 }, previousSaved);
    };
  }
  const like = trpc.social.setLike.useMutation({
    onMutate: async ({ postId, liked }) =>
      optimistic(postId, (post) => applyOptimisticLike(post, liked)),
    onError: (_error, _input, rollback) => rollback?.(),
    onSettled: invalidate,
  });
  const save = trpc.social.setSaved.useMutation({
    onMutate: async ({ postId, saved }) =>
      optimistic(postId, (post) => applyOptimisticSave(post, saved)),
    onError: (_error, _input, rollback) => rollback?.(),
    onSettled: invalidate,
  });
  return {
    query,
    posts: query.data?.pages.flatMap((page) => page.items) ?? [],
    create,
    update,
    remove,
    like,
    save,
  };
}
