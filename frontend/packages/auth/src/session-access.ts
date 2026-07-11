import {
  createAccessContext,
  isAccessTier,
  isPlatformRole,
  type AccessContext,
  type AccessTier,
  type PlatformRole,
} from "./access";

export const COACH_CONNECT_SESSION_CLAIM = "coach_connect";

export interface VerifiedSessionInput {
  userId: string | null;
  sessionClaims: unknown;
}

interface AccessClaim {
  role: PlatformRole;
  tier: AccessTier;
}

const safeAccessClaim: AccessClaim = { role: "user", tier: "free" };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function readAccessClaim(sessionClaims: unknown): AccessClaim | null {
  if (!isRecord(sessionClaims)) return null;

  const accessClaim = sessionClaims[COACH_CONNECT_SESSION_CLAIM];
  if (!isRecord(accessClaim)) return null;
  if (!isPlatformRole(accessClaim.role) || !isAccessTier(accessClaim.tier)) {
    return null;
  }

  return {
    role: accessClaim.role,
    tier: accessClaim.tier,
  };
}

export function resolveAccessFromSession({
  userId,
  sessionClaims,
}: VerifiedSessionInput): AccessContext {
  if (!userId) return createAccessContext({ userId: null });

  const claim = readAccessClaim(sessionClaims) ?? safeAccessClaim;
  return createAccessContext({
    userId,
    role: claim.role,
    tier: claim.tier,
  });
}
