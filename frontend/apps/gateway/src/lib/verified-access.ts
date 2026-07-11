import "server-only";

import { createAccessContext } from "@coach-connect/auth/access";
import { applyVerifiedEmailAdminOverride } from "@coach-connect/auth/email-admin";
import { resolveAccessFromSession } from "@coach-connect/auth/session-access";
import { auth, currentUser } from "@clerk/nextjs/server";

const permanentAdminEmails = ["source.sensei1205@gmail.com"] as const;

async function getVerifiedPrimaryEmail(userId: string): Promise<string | null> {
  try {
    const user = await currentUser();
    const primaryEmail = user?.primaryEmailAddress;

    return user?.id === userId &&
      primaryEmail?.verification?.status === "verified"
      ? primaryEmail.emailAddress
      : null;
  } catch {
    return null;
  }
}

export async function getVerifiedAccess() {
  const { userId, sessionClaims } = await auth();
  const access = resolveAccessFromSession({ userId, sessionClaims });

  if (!access.isAuthenticated) {
    return access;
  }

  const verifiedPrimaryEmail = await getVerifiedPrimaryEmail(access.userId);

  return applyVerifiedEmailAdminOverride(
    access,
    verifiedPrimaryEmail,
    permanentAdminEmails,
  );
}

export async function getRequestAccess() {
  const isClerkConfigured = Boolean(
    process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY &&
    process.env.CLERK_SECRET_KEY,
  );

  return isClerkConfigured
    ? getVerifiedAccess()
    : createAccessContext({ userId: null });
}
