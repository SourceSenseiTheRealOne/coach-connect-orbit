import { describe, expect, it } from "vitest";

import { createAccessContext } from "./access";
import { FeatureGate } from "./feature-gate";

describe("FeatureGate", () => {
  it("returns children when the access requirement passes", () => {
    expect(
      FeatureGate({
        access: createAccessContext({ userId: "pro", tier: "pro" }),
        requirement: { kind: "minimum-tier", tier: "pro" },
        children: "premium feature",
        fallback: "upgrade",
      }),
    ).toBe("premium feature");
  });

  it("returns the fallback when the access requirement fails", () => {
    expect(
      FeatureGate({
        access: createAccessContext({ userId: "free", tier: "free" }),
        requirement: { kind: "minimum-tier", tier: "pro" },
        children: "premium feature",
        fallback: "upgrade",
      }),
    ).toBe("upgrade");
  });

  it("hides denied content by default", () => {
    expect(
      FeatureGate({
        access: createAccessContext({ userId: "user", role: "user" }),
        requirement: { kind: "admin-only" },
        children: "moderation",
      }),
    ).toBeNull();
  });
});
