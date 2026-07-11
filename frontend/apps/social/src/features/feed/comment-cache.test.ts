import { describe, expect, it, vi } from "vitest";
import { refreshCommentMutationCaches } from "./comment-cache";

describe("refreshCommentMutationCaches", () => {
  it("refreshes comments and both post-list caches", async () => {
    const refreshComments = vi.fn().mockResolvedValue(undefined);
    const refreshLatest = vi.fn().mockResolvedValue(undefined);
    const refreshSaved = vi.fn().mockResolvedValue(undefined);

    await refreshCommentMutationCaches(
      refreshComments,
      refreshLatest,
      refreshSaved,
    );

    expect(refreshComments).toHaveBeenCalledOnce();
    expect(refreshLatest).toHaveBeenCalledOnce();
    expect(refreshSaved).toHaveBeenCalledOnce();
  });
});
