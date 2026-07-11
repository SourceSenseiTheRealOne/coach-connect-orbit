import { describe, expect, it } from "vitest";

import { createTierCatalog } from "./tier-catalog";

describe("createTierCatalog", () => {
  it("contains exactly free, pro, and elite with mock monthly EUR prices", () => {
    expect(createTierCatalog()).toEqual([
      {
        tier: "free",
        displayName: "Free",
        price: { amountMinor: 0, currency: "EUR", interval: "month" },
        billingReference: { provider: "mock" },
      },
      {
        tier: "pro",
        displayName: "Pro",
        price: { amountMinor: 999, currency: "EUR", interval: "month" },
        billingReference: { provider: "mock" },
      },
      {
        tier: "elite",
        displayName: "Elite",
        price: { amountMinor: 1999, currency: "EUR", interval: "month" },
        billingReference: { provider: "mock" },
      },
    ]);
  });

  it("injects future Stripe display prices without changing tier slugs", () => {
    const catalog = createTierCatalog({
      pro: {
        amountMinor: 1299,
        currency: "EUR",
        stripePriceId: "price_pro",
      },
      elite: {
        amountMinor: 2499,
        currency: "EUR",
        stripePriceId: "price_elite",
      },
    });

    expect(catalog.map(({ tier }) => tier)).toEqual(["free", "pro", "elite"]);
    expect(catalog[1]).toMatchObject({
      price: { amountMinor: 1299 },
      billingReference: { provider: "stripe", priceId: "price_pro" },
    });
    expect(catalog[2]).toMatchObject({
      price: { amountMinor: 2499 },
      billingReference: { provider: "stripe", priceId: "price_elite" },
    });
  });

  it("rejects invalid paid prices and blank Stripe Price IDs", () => {
    expect(() =>
      createTierCatalog({ pro: { amountMinor: 0, currency: "EUR" } }),
    ).toThrow("pro amountMinor must be a positive integer");
    expect(() =>
      createTierCatalog({
        elite: { amountMinor: 2000, currency: "EUR", stripePriceId: " " },
      }),
    ).toThrow("elite stripePriceId must not be blank");
  });
});
