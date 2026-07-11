# `@coach-connect/auth`

Shared Clerk provider, verified access-policy primitives, feature gates, and tier catalog for every frontend zone.

## Access model

Access is intentionally split into independent values:

- verified Clerk `userId`;
- platform role: `user` or `admin`;
- subscription tier: `free`, `pro`, or `elite`;
- a feature requirement: authenticated, admin-only, or minimum tier.

Admin is not a paid tier. A free admin only passes a paid-tier requirement when that requirement explicitly sets `allowAdminOverride: true`. `FeatureGate` and navigation filtering are presentation-only. Protected tRPC procedures and Go use cases must enforce the same business policy before returning data or mutating state.

## Clerk session claim used during the mock phase

In Clerk Dashboard → Sessions → Customize session token, add this claim:

```json
{
  "coach_connect": "{{user.public_metadata.coach_connect}}"
}
```

The backend-managed Clerk public metadata shape is:

```json
{
  "coach_connect": {
    "role": "user",
    "tier": "free"
  }
}
```

Do not let clients update this metadata. `resolveAccessFromSession()` accepts Clerk's already-verified server-side `sessionClaims`; the role and tier are validated atomically, so a missing or malformed field makes the entire entitlement claim fail closed to `user` + `free`. `createAccessContext()` enforces the same atomic rule: omitting both values gives an authenticated user the default `user` + `free` access, while supplying only one or mixing a valid value with an invalid value fails the whole pair closed. After changing metadata during development, reload the Clerk session so a fresh token contains the update.

The gateway separately checks Clerk's server-fetched, verified primary email against the permanent owner-admin allowlist. A match grants only the independent platform `admin` role and preserves the resolved subscription tier; unverified, mismatched, or missing emails never elevate access.

The claim is temporary entitlement scaffolding. Future Stripe subscription events will update server-owned subscription state behind a resolver with the same stable tier slugs. Security-sensitive server operations must not depend only on UI-visible claim state.

## Feature checks

```ts
import { canAccess } from "@coach-connect/auth/access";

const canUseProFeature = canAccess(access, {
  kind: "minimum-tier",
  tier: "pro",
});

const canModerate = canAccess(access, { kind: "admin-only" });
```

Use `requireAccess()` at trusted TypeScript server boundaries. Go still owns durable resource authorization.

## Tier catalog

`createTierCatalog()` currently returns mock monthly prices:

- Free: EUR 0
- Pro: EUR 9.99
- Elite: EUR 19.99

Later, inject final amounts and Stripe Price IDs through `createTierCatalog({ pro, elite })`. Never infer entitlement from displayed amounts and never trust a browser-supplied Stripe Price ID.
