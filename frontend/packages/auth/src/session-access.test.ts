import { describe, expect, it } from "vitest";

import { resolveAccessFromSession } from "./session-access";

describe("resolveAccessFromSession", () => {
  it("returns guest access without a Clerk user ID", () => {
    expect(
      resolveAccessFromSession({
        userId: null,
        sessionClaims: { coach_connect: { role: "admin", tier: "elite" } },
      }),
    ).toMatchObject({ state: "guest", role: null, tier: null });
  });

  it("reads role and tier only from the signed Coach Connect claim", () => {
    expect(
      resolveAccessFromSession({
        userId: "user_admin",
        sessionClaims: {
          coach_connect: { role: "admin", tier: "elite" },
          role: "user",
          tier: "free",
        },
      }),
    ).toMatchObject({
      state: "authenticated",
      userId: "user_admin",
      role: "admin",
      tier: "elite",
    });
  });

  it("fails closed when the custom claim has an invalid shape", () => {
    expect(
      resolveAccessFromSession({
        userId: "user_123",
        sessionClaims: { coach_connect: "admin" },
      }),
    ).toMatchObject({ role: "user", tier: "free" });
  });

  it.each([
    { role: "admin", tier: undefined },
    { role: "admin", tier: "enterprise" },
    { role: undefined, tier: "elite" },
    { role: "owner", tier: "elite" },
  ])("fails closed atomically for mixed-validity claim %#", (coachConnect) => {
    expect(
      resolveAccessFromSession({
        userId: "user_mixed",
        sessionClaims: { coach_connect: coachConnect },
      }),
    ).toMatchObject({
      role: "user",
      tier: "free",
      isAdmin: false,
      isPaid: false,
    });
  });
});
