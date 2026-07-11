import type { Post } from "@coach-connect/go-api-client";
import { describe, expect, it } from "vitest";
import { applyOptimisticLike, applyOptimisticSave } from "./optimistic";

const post = {
  id: "p",
  viewerHasLiked: false,
  viewerHasSaved: false,
  likeCount: 0,
} as Post;

describe("optimistic feed state", () => {
  it("is idempotent and never produces a negative like count", () => {
    const liked = applyOptimisticLike(post, true);
    expect(applyOptimisticLike(liked, true)).toBe(liked);
    expect(applyOptimisticLike(post, false)).toBe(post);
    expect(
      applyOptimisticLike({ ...post, viewerHasLiked: true }, false).likeCount,
    ).toBe(0);
  });

  it("preserves the original snapshot for rollback", () => {
    const optimistic = applyOptimisticSave(post, true);
    expect(optimistic.viewerHasSaved).toBe(true);
    expect(post.viewerHasSaved).toBe(false);
  });
});
