import { describe, expect, it } from "vitest";

import { createAccessContext } from "./access";
import { applyVerifiedEmailAdminOverride } from "./email-admin";

const ownerEmail = "source.sensei1205@gmail.com";

describe("applyVerifiedEmailAdminOverride", () => {
  it("grants the admin role to an authenticated matching verified email", () => {
    const access = createAccessContext({ userId: "owner" });

    expect(
      applyVerifiedEmailAdminOverride(access, ownerEmail, [ownerEmail]),
    ).toMatchObject({
      role: "admin",
      tier: "free",
      isAdmin: true,
      isPaid: false,
    });
  });

  it("matches email addresses case-insensitively and preserves the tier", () => {
    const eliteAccess = createAccessContext({
      userId: "owner",
      role: "user",
      tier: "elite",
    });

    expect(
      applyVerifiedEmailAdminOverride(eliteAccess, ownerEmail.toUpperCase(), [
        ` ${ownerEmail} `,
      ]),
    ).toMatchObject({ role: "admin", tier: "elite", isPaid: true });
  });

  it("does not elevate guests or unrelated verified emails", () => {
    const guest = createAccessContext({ userId: null });
    const user = createAccessContext({ userId: "user" });

    expect(
      applyVerifiedEmailAdminOverride(guest, ownerEmail, [ownerEmail]),
    ).toBe(guest);
    expect(
      applyVerifiedEmailAdminOverride(user, "other@example.com", [ownerEmail]),
    ).toBe(user);
  });

  it("does not elevate when no verified email is available", () => {
    const user = createAccessContext({ userId: "user" });

    expect(applyVerifiedEmailAdminOverride(user, null, [ownerEmail])).toBe(
      user,
    );
  });
});
