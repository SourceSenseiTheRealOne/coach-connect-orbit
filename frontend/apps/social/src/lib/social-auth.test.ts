import { describe, expect, it } from "vitest";
import { socialSignInURL } from "./social-auth";

describe("socialSignInURL", () => {
  it("redirects standalone social routes to the gateway sign-in page", () => {
    expect(
      socialSignInURL(
        "http://social:3001/feed?view=saved",
        "https://coach.example.com",
      ).toString(),
    ).toBe(
      "https://coach.example.com/sign-in?redirect_url=%2Ffeed%3Fview%3Dsaved",
    );
  });

  it("rejects gateway values that are not bare HTTP origins", () => {
    expect(() =>
      socialSignInURL(
        "http://social:3001/feed",
        "https://user:password@coach.example.com/private",
      ),
    ).toThrow("GATEWAY_PUBLIC_ORIGIN must be an HTTP(S) origin");
  });
});
