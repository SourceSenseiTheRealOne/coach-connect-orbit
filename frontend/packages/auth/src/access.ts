export const ACCESS_TIERS = ["free", "pro", "elite"] as const;

export type AccessTier = (typeof ACCESS_TIERS)[number];
export type PlatformRole = "user" | "admin";

export interface GuestAccessContext {
  state: "guest";
  isAuthenticated: false;
  isAdmin: false;
  isPaid: false;
  role: null;
  tier: null;
  userId: null;
}

export interface AuthenticatedAccessContext {
  state: "authenticated";
  isAuthenticated: true;
  isAdmin: boolean;
  isPaid: boolean;
  role: PlatformRole;
  tier: AccessTier;
  userId: string;
}

export type AccessContext = GuestAccessContext | AuthenticatedAccessContext;

export type AccessRequirement =
  | { kind: "authenticated" }
  | { kind: "admin-only" }
  | {
      kind: "minimum-tier";
      tier: AccessTier;
      allowAdminOverride?: boolean;
    };

export interface AccessContextInput {
  userId: string | null;
  role?: unknown;
  tier?: unknown;
}

const tierRank = {
  free: 0,
  pro: 1,
  elite: 2,
} as const satisfies Record<AccessTier, number>;

export function isAccessTier(value: unknown): value is AccessTier {
  return (
    typeof value === "string" && ACCESS_TIERS.includes(value as AccessTier)
  );
}

export function isPlatformRole(value: unknown): value is PlatformRole {
  return value === "user" || value === "admin";
}

function normalizeRole(value: unknown): PlatformRole {
  return value === "admin" ? "admin" : "user";
}

function normalizeTier(value: unknown): AccessTier {
  return isAccessTier(value) ? value : "free";
}

export function createAccessContext(input: AccessContextInput): AccessContext {
  if (!input.userId) {
    return {
      state: "guest",
      isAuthenticated: false,
      isAdmin: false,
      isPaid: false,
      role: null,
      tier: null,
      userId: null,
    };
  }

  const role = normalizeRole(input.role);
  const tier = normalizeTier(input.tier);

  return {
    state: "authenticated",
    isAuthenticated: true,
    isAdmin: role === "admin",
    isPaid: tier !== "free",
    role,
    tier,
    userId: input.userId,
  };
}

export function canAccess(
  access: AccessContext,
  requirement: AccessRequirement,
): boolean {
  if (!access.isAuthenticated) return false;

  switch (requirement.kind) {
    case "authenticated":
      return true;
    case "admin-only":
      return access.isAdmin;
    case "minimum-tier":
      return (
        (requirement.allowAdminOverride === true && access.isAdmin) ||
        tierRank[access.tier] >= tierRank[requirement.tier]
      );
  }
}

export type AccessDeniedCode = "unauthenticated" | "forbidden";

export class AccessDeniedError extends Error {
  readonly code: AccessDeniedCode;

  constructor(code: AccessDeniedCode) {
    super(code);
    this.name = "AccessDeniedError";
    this.code = code;
  }
}

export function requireAccess(
  access: AccessContext,
  requirement: AccessRequirement,
): AuthenticatedAccessContext {
  if (!access.isAuthenticated) {
    throw new AccessDeniedError("unauthenticated");
  }

  if (!canAccess(access, requirement)) {
    throw new AccessDeniedError("forbidden");
  }

  return access;
}
