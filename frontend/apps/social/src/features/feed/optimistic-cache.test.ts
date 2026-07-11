import { describe, expect, it, vi } from "vitest";
import { cancelFeedQueriesBeforeSnapshot } from "./optimistic-cache";

describe("cancelFeedQueriesBeforeSnapshot", () => {
  it("waits for both feed queries before mutation snapshots proceed", async () => {
    let releaseLatest!: () => void;
    let releaseSaved!: () => void;
    const latest = new Promise<void>((resolve) => {
      releaseLatest = resolve;
    });
    const saved = new Promise<void>((resolve) => {
      releaseSaved = resolve;
    });
    const cancelLatest = vi.fn(() => latest);
    const cancelSaved = vi.fn(() => saved);
    let finished = false;

    const cancellation = cancelFeedQueriesBeforeSnapshot(
      cancelLatest,
      cancelSaved,
    ).then(() => {
      finished = true;
    });

    await Promise.resolve();
    expect(cancelLatest).toHaveBeenCalledOnce();
    expect(cancelSaved).toHaveBeenCalledOnce();
    expect(finished).toBe(false);

    releaseLatest();
    await Promise.resolve();
    expect(finished).toBe(false);

    releaseSaved();
    await cancellation;
    expect(finished).toBe(true);
  });
});
