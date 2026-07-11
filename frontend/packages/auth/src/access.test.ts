import { describe, expect, it } from "vitest";

import {
  AccessDeniedError,
  canAccess,
  createAccessContext,
  requireAccess,
  type AccessRequirement,
} from "./access";

const authenticated: AccessRequirement = { kind: "authenticated" };
const proFeature: AccessRequirement = {
  kind: "minimum-tier",
  tier: "pro",
};
const proFeatureWithAdminOverride: AccessRequirement = {
  kind: "minimum-tier",
  tier: "pro",
  allowAdminOverride: true,
};
const adminFeature: AccessRequirement = { kind: "admin-only" };

describe("createAccessContext", () => {
  it("represents a missing user as an unauthenticated guest", () => {
    expect(createAccessContext({ userId: null })).toEqual({
      state: "guest",
      isAuthenticated: false,
      isAdmin: false,
      isPaid: false,
      role: null,
      tier: null,
      userId: null,
    });
  });

  it("fails malformed role and tier values closed to user and free", () => {
    expect(
      createAccessContext({
        userId: "user_123",
        role: "super-admin",
        tier: "enterprise",
      }),
    ).toMatchObject({
      state: "authenticated",
      role: "user",
      tier: "free",
      isAdmin: false,
      isPaid: false,
    });
  });

  it.each([
    { role: "admin", tier: undefined },
    { role: "admin", tier: "enterprise" },
    { role: undefined, tier: "elite" },
    { role: "owner", tier: "elite" },
  ])("fails mixed-validity values closed atomically %#", ({ role, tier }) => {
    expect(
      createAccessContext({ userId: "user_mixed", role, tier }),
    ).toMatchObject({
      role: "user",
      tier: "free",
      isAdmin: false,
      isPaid: false,
    });
  });

  it.each(["pro", "elite"] as const)("marks the %s tier as paid", (tier) => {
    expect(
      createAccessContext({ userId: "user_123", role: "user", tier }).isPaid,
    ).toBe(true);
  });
});

describe("canAccess", () => {
  it("requires a verified user for authenticated features", () => {
    expect(
      canAccess(createAccessContext({ userId: null }), authenticated),
    ).toBe(false);
    expect(
      canAccess(createAccessContext({ userId: "user_123" }), authenticated),
    ).toBe(true);
  });

  it("uses the free < pro < elite tier hierarchy", () => {
    expect(
      canAccess(
        createAccessContext({ userId: "free", role: "user", tier: "free" }),
        proFeature,
      ),
    ).toBe(false);
    expect(
      canAccess(
        createAccessContext({ userId: "pro", role: "user", tier: "pro" }),
        proFeature,
      ),
    ).toBe(true);
    expect(
      canAccess(
        createAccessContext({ userId: "elite", role: "user", tier: "elite" }),
        proFeature,
      ),
    ).toBe(true);
  });

  it("does not treat admin as a paid tier unless the feature opts in", () => {
    const freeAdmin = createAccessContext({
      userId: "admin",
      role: "admin",
      tier: "free",
    });

    expect(canAccess(freeAdmin, proFeature)).toBe(false);
    expect(canAccess(freeAdmin, proFeatureWithAdminOverride)).toBe(true);
    expect(canAccess(freeAdmin, adminFeature)).toBe(true);
  });

  it("does not grant admin-only access to an elite non-admin", () => {
    const eliteUser = createAccessContext({
      userId: "elite",
      role: "user",
      tier: "elite",
    });

    expect(canAccess(eliteUser, adminFeature)).toBe(false);
  });
});

describe("requireAccess", () => {
  it("distinguishes unauthenticated from forbidden access", () => {
    expect(() =>
      requireAccess(createAccessContext({ userId: null }), authenticated),
    ).toThrowError(new AccessDeniedError("unauthenticated"));

    expect(() =>
      requireAccess(
        createAccessContext({ userId: "free", role: "user", tier: "free" }),
        proFeature,
      ),
    ).toThrowError(new AccessDeniedError("forbidden"));
  });
});
