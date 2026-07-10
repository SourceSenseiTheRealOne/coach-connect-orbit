"use client";

import { ClerkProvider } from "@clerk/nextjs";
import type { ReactNode } from "react";

export interface CoachConnectAuthProviderProps {
  children: ReactNode;
  publishableKey: string | undefined;
}

export function CoachConnectAuthProvider({
  children,
  publishableKey,
}: CoachConnectAuthProviderProps) {
  if (!publishableKey) {
    return children;
  }

  return (
    <ClerkProvider publishableKey={publishableKey}>{children}</ClerkProvider>
  );
}
