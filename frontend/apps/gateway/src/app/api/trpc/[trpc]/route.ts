import "server-only";

import { localServiceOrigins } from "@coach-connect/config";
import { createGoApiClient } from "@coach-connect/go-api-client";
import { appRouter } from "@coach-connect/trpc-contract";
import { fetchRequestHandler } from "@trpc/server/adapters/fetch";

import { getRequestAuth } from "../../../../lib/verified-access";

const api = createGoApiClient({
  baseUrl: process.env.GO_API_URL ?? localServiceOrigins.goApi,
});

async function handler(request: Request) {
  const { access, bearerToken, identity } = await getRequestAuth();

  return fetchRequestHandler({
    endpoint: "/api/trpc",
    req: request,
    router: appRouter,
    createContext: () => ({ access, api, bearerToken, identity }),
    onError: ({ error, path }) => {
      if (process.env.NODE_ENV === "development") {
        console.error(`tRPC failure on ${path ?? "unknown procedure"}`, error);
      }
    },
  });
}

export { handler as GET, handler as POST };
