import type { components } from "./generated/schema";

export type HealthResponse = components["schemas"]["HealthResponse"];

export interface GoApiClient {
  health(): Promise<HealthResponse>;
}

export interface GoApiClientOptions {
  baseUrl: string;
  fetchImplementation?: typeof fetch;
}

function parseHealthResponse(value: unknown): HealthResponse {
  if (
    typeof value !== "object" ||
    value === null ||
    !("status" in value) ||
    value.status !== "ok" ||
    !("service" in value) ||
    value.service !== "coach-connect-api"
  ) {
    throw new TypeError("Go API returned an invalid health response");
  }

  return { status: value.status, service: value.service };
}

export function createGoApiClient({
  baseUrl,
  fetchImplementation = fetch,
}: GoApiClientOptions): GoApiClient {
  const normalizedBaseUrl = baseUrl.replace(/\/$/, "");

  return {
    async health() {
      const response = await fetchImplementation(
        `${normalizedBaseUrl}/api/v1/health`,
        {
          method: "GET",
          headers: { accept: "application/json" },
        },
      );

      if (!response.ok) {
        throw new Error(`Go API health request failed with ${response.status}`);
      }

      return parseHealthResponse(await response.json());
    },
  };
}
