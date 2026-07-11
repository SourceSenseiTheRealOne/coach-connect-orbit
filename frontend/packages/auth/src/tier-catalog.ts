import type { AccessTier } from "./access";

export interface TierPrice {
  amountMinor: number;
  currency: string;
  interval: "month";
}

export type BillingReference =
  { provider: "mock" } | { provider: "stripe"; priceId: string };

export interface TierPlan {
  tier: AccessTier;
  displayName: string;
  price: TierPrice;
  billingReference: BillingReference;
}

export interface PaidTierCatalogOverride {
  amountMinor: number;
  currency: string;
  stripePriceId?: string;
}

export interface TierCatalogOverrides {
  pro?: PaidTierCatalogOverride;
  elite?: PaidTierCatalogOverride;
}

const mockPaidPlans = {
  pro: { amountMinor: 999, currency: "EUR" },
  elite: { amountMinor: 1999, currency: "EUR" },
} as const satisfies Record<
  "pro" | "elite",
  Omit<PaidTierCatalogOverride, "stripePriceId">
>;

function createPaidPlan(
  tier: "pro" | "elite",
  displayName: "Pro" | "Elite",
  override: PaidTierCatalogOverride | undefined,
): TierPlan {
  const price = override ?? mockPaidPlans[tier];

  if (!Number.isInteger(price.amountMinor) || price.amountMinor <= 0) {
    throw new Error(`${tier} amountMinor must be a positive integer`);
  }
  if (!/^[A-Z]{3}$/.test(price.currency)) {
    throw new Error(`${tier} currency must be an uppercase ISO-4217 code`);
  }

  const stripePriceId = override?.stripePriceId;
  if (stripePriceId !== undefined && stripePriceId.trim().length === 0) {
    throw new Error(`${tier} stripePriceId must not be blank`);
  }

  return {
    tier,
    displayName,
    price: {
      amountMinor: price.amountMinor,
      currency: price.currency,
      interval: "month",
    },
    billingReference:
      stripePriceId === undefined
        ? { provider: "mock" }
        : { provider: "stripe", priceId: stripePriceId },
  };
}

export function createTierCatalog(
  overrides: TierCatalogOverrides = {},
): readonly [TierPlan, TierPlan, TierPlan] {
  return [
    {
      tier: "free",
      displayName: "Free",
      price: { amountMinor: 0, currency: "EUR", interval: "month" },
      billingReference: { provider: "mock" },
    },
    createPaidPlan("pro", "Pro", overrides.pro),
    createPaidPlan("elite", "Elite", overrides.elite),
  ];
}
