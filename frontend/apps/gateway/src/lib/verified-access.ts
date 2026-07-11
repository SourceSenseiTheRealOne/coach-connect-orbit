import "server-only";

import { createAccessContext } from "@coach-connect/auth/access";
import { resolveAccessFromSession } from "@coach-connect/auth/session-access";
import { auth } from "@clerk/nextjs/server";

export async function getVerifiedAccess() {
  const { userId, sessionClaims } = await auth();

  return resolveAccessFromSession({ userId, sessionClaims });
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
