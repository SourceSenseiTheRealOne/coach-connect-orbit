import type { components } from "./generated/schema";

export type HealthResponse = components["schemas"]["HealthResponse"];
export type Post = components["schemas"]["Post"];
export type FeedPage = components["schemas"]["FeedPage"];
export type Comment = components["schemas"]["Comment"];
export type CommentPage = components["schemas"]["CommentPage"];

export interface AuthenticatedRequestOptions {
  bearerToken: string;
  signal?: AbortSignal;
}

export interface PageRequestOptions extends AuthenticatedRequestOptions {
  cursor?: string | undefined;
  limit?: number | undefined;
}

export interface GoApiClient {
  health(): Promise<HealthResponse>;
  listFeed(options: PageRequestOptions): Promise<FeedPage>;
  listSavedPosts(options: PageRequestOptions): Promise<FeedPage>;
  getPost(postId: string, options: AuthenticatedRequestOptions): Promise<Post>;
  createPost(body: string, options: AuthenticatedRequestOptions): Promise<Post>;
  updatePost(
    postId: string,
    body: string,
    options: AuthenticatedRequestOptions,
  ): Promise<Post>;
  deletePost(
    postId: string,
    options: AuthenticatedRequestOptions,
  ): Promise<void>;
  setPostLike(
    postId: string,
    liked: boolean,
    options: AuthenticatedRequestOptions,
  ): Promise<Post>;
  setSavedPost(
    postId: string,
    saved: boolean,
    options: AuthenticatedRequestOptions,
  ): Promise<Post>;
  listComments(
    postId: string,
    options: PageRequestOptions,
  ): Promise<CommentPage>;
  createComment(
    postId: string,
    body: string,
    options: AuthenticatedRequestOptions,
  ): Promise<Comment>;
  updateComment(
    commentId: string,
    body: string,
    options: AuthenticatedRequestOptions,
  ): Promise<Comment>;
  deleteComment(
    commentId: string,
    options: AuthenticatedRequestOptions,
  ): Promise<void>;
}

export interface GoApiClientOptions {
  baseUrl: string;
  fetchImplementation?: typeof fetch;
  timeoutMs?: number;
}

export class GoApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string) {
    super(`Go API request failed with ${status}: ${code}`);
    this.name = "GoApiError";
    this.status = status;
    this.code = code;
  }
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

function authHeaders(options: AuthenticatedRequestOptions) {
  return {
    accept: "application/json",
    authorization: `Bearer ${options.bearerToken}`,
  };
}

function pageSearch(options: PageRequestOptions) {
  const search = new URLSearchParams();
  if (options.cursor) {
    search.set("cursor", options.cursor);
  }
  if (options.limit !== undefined) {
    search.set("limit", String(options.limit));
  }
  const encoded = search.toString();
  return encoded ? `?${encoded}` : "";
}

async function parseJson<T>(response: Response): Promise<T> {
  if (!response.ok) {
    let code = "request_failed";
    try {
      const payload = (await response.json()) as { code?: unknown };
      if (typeof payload.code === "string") {
        code = payload.code;
      }
    } catch {
      code = "invalid_error_response";
    }
    throw new GoApiError(response.status, code);
  }

  return (await response.json()) as T;
}

export function createGoApiClient({
  baseUrl,
  fetchImplementation = fetch,
  timeoutMs = 8_000,
}: GoApiClientOptions): GoApiClient {
  const normalizedBaseUrl = baseUrl.replace(/\/$/, "");

  async function requestJson<T>(path: string, init: RequestInit): Promise<T> {
    const timeout = AbortSignal.timeout(timeoutMs);
    const signal = init.signal
      ? AbortSignal.any([init.signal, timeout])
      : timeout;
    try {
      return parseJson<T>(
        await fetchImplementation(`${normalizedBaseUrl}${path}`, {
          ...init,
          signal,
        }),
      );
    } catch (error) {
      if (error instanceof GoApiError) throw error;
      if (timeout.aborted) throw new GoApiError(504, "timeout");
      if (signal.aborted) throw new GoApiError(499, "cancelled");
      throw new GoApiError(503, "service_unavailable");
    }
  }

  async function requestVoid(path: string, init: RequestInit): Promise<void> {
    const timeout = AbortSignal.timeout(timeoutMs);
    const signal = init.signal
      ? AbortSignal.any([init.signal, timeout])
      : timeout;
    try {
      const response = await fetchImplementation(
        `${normalizedBaseUrl}${path}`,
        {
          ...init,
          signal,
        },
      );
      if (!response.ok) {
        await parseJson<never>(response);
      }
    } catch (error) {
      if (error instanceof GoApiError) throw error;
      if (timeout.aborted) throw new GoApiError(504, "timeout");
      if (signal.aborted) throw new GoApiError(499, "cancelled");
      throw new GoApiError(503, "service_unavailable");
    }
  }

  function optionalSignal(
    signal: AbortSignal | undefined,
  ): Pick<RequestInit, "signal"> | Record<string, never> {
    return signal === undefined ? {} : { signal };
  }

  function bodyRequest(body: string) {
    return JSON.stringify({ body });
  }

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
        throw new GoApiError(response.status, "health_failed");
      }

      return parseHealthResponse(await response.json());
    },
    listFeed(options) {
      return requestJson<FeedPage>(`/api/v1/feed${pageSearch(options)}`, {
        method: "GET",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
    listSavedPosts(options) {
      return requestJson<FeedPage>(
        `/api/v1/saved-posts${pageSearch(options)}`,
        {
          method: "GET",
          headers: authHeaders(options),
          ...optionalSignal(options.signal),
        },
      );
    },
    getPost(postId, options) {
      return requestJson<Post>(`/api/v1/posts/${postId}`, {
        method: "GET",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
    createPost(body, options) {
      return requestJson<Post>("/api/v1/posts", {
        method: "POST",
        headers: {
          ...authHeaders(options),
          "content-type": "application/json",
        },
        body: bodyRequest(body),
        ...optionalSignal(options.signal),
      });
    },
    updatePost(postId, body, options) {
      return requestJson<Post>(`/api/v1/posts/${postId}`, {
        method: "PATCH",
        headers: {
          ...authHeaders(options),
          "content-type": "application/json",
        },
        body: bodyRequest(body),
        ...optionalSignal(options.signal),
      });
    },
    async deletePost(postId, options) {
      await requestVoid(`/api/v1/posts/${postId}`, {
        method: "DELETE",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
    setPostLike(postId, liked, options) {
      return requestJson<Post>(`/api/v1/posts/${postId}/like`, {
        method: liked ? "PUT" : "DELETE",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
    setSavedPost(postId, saved, options) {
      return requestJson<Post>(`/api/v1/posts/${postId}/save`, {
        method: saved ? "PUT" : "DELETE",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
    listComments(postId, options) {
      return requestJson<CommentPage>(
        `/api/v1/posts/${postId}/comments${pageSearch(options)}`,
        {
          method: "GET",
          headers: authHeaders(options),
          ...optionalSignal(options.signal),
        },
      );
    },
    createComment(postId, body, options) {
      return requestJson<Comment>(`/api/v1/posts/${postId}/comments`, {
        method: "POST",
        headers: {
          ...authHeaders(options),
          "content-type": "application/json",
        },
        body: bodyRequest(body),
        ...optionalSignal(options.signal),
      });
    },
    updateComment(commentId, body, options) {
      return requestJson<Comment>(`/api/v1/comments/${commentId}`, {
        method: "PATCH",
        headers: {
          ...authHeaders(options),
          "content-type": "application/json",
        },
        body: bodyRequest(body),
        ...optionalSignal(options.signal),
      });
    },
    async deleteComment(commentId, options) {
      await requestVoid(`/api/v1/comments/${commentId}`, {
        method: "DELETE",
        headers: authHeaders(options),
        ...optionalSignal(options.signal),
      });
    },
  };
}
