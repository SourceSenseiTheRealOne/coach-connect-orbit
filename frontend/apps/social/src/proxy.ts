import { clerkMiddleware, createRouteMatcher } from "@clerk/nextjs/server";
import {
  type NextFetchEvent,
  type NextRequest,
  NextResponse,
} from "next/server";
import { socialSignInURL } from "./lib/social-auth";

const isProtectedRoute = createRouteMatcher(["/feed(.*)"]);
const isClerkConfigured = Boolean(
  process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY && process.env.CLERK_SECRET_KEY,
);

function redirectToGatewaySignIn(request: NextRequest) {
  return NextResponse.redirect(
    socialSignInURL(request.url, process.env.GATEWAY_PUBLIC_ORIGIN),
  );
}

const clerkProxy = clerkMiddleware(async (auth, request) => {
  if (!isProtectedRoute(request)) return;
  const { userId } = await auth();
  if (!userId) return redirectToGatewaySignIn(request);
});

export default function proxy(request: NextRequest, event: NextFetchEvent) {
  if (!isProtectedRoute(request)) return NextResponse.next();
  if (!isClerkConfigured) return redirectToGatewaySignIn(request);
  return clerkProxy(request, event);
}

export const config = {
  matcher: ["/feed(.*)"],
};
