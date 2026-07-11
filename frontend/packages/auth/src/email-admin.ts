import { createAccessContext, type AccessContext } from "./access";

function normalizeEmail(value: string): string {
  return value.trim().toLowerCase();
}

export function applyVerifiedEmailAdminOverride(
  access: AccessContext,
  verifiedEmail: string | null,
  adminEmails: readonly string[],
): AccessContext {
  if (!access.isAuthenticated || !verifiedEmail) {
    return access;
  }

  const normalizedEmail = normalizeEmail(verifiedEmail);
  const isAdminEmail = adminEmails.some(
    (adminEmail) => normalizeEmail(adminEmail) === normalizedEmail,
  );

  if (!isAdminEmail) {
    return access;
  }

  return createAccessContext({
    userId: access.userId,
    role: "admin",
    tier: access.tier,
  });
}
