import type { ReactNode } from "react";

import {
  canAccess,
  type AccessContext,
  type AccessRequirement,
} from "./access";

export interface FeatureGateProps {
  access: AccessContext;
  requirement: AccessRequirement;
  children: ReactNode;
  fallback?: ReactNode;
}

export function FeatureGate({
  access,
  requirement,
  children,
  fallback = null,
}: FeatureGateProps): ReactNode {
  return canAccess(access, requirement) ? children : fallback;
}
