// Server-component fetch helper. Calls the API Gateway directly by absolute URL
// (the /api/* rewrite in next.config.ts exists for the browser only — server code
// doesn't need to hop through :3000). Returns null on any failure so pages can
// render a graceful empty/error state instead of crashing when a backend is down.
import { ApiError } from "./apiError";
import { GATEWAY_URL } from "./gatewayUrl";

export async function serverApi<T>(
  path: string,
  init: RequestInit = {},
): Promise<{ data: T | null; error: ApiError | null }> {
  try {
    const res = await fetch(`${GATEWAY_URL}${path}`, {
      ...init,
      cache: init.cache ?? "no-store",
    });
    if (res.status === 204) return { data: undefined as T, error: null };
    const body = await res.json().catch(() => null);
    if (!res.ok) {
      return { data: null, error: new ApiError(res.status, body ?? { error: "unknown", message: "Request failed" }) };
    }
    return { data: body as T, error: null };
  } catch {
    return { data: null, error: new ApiError(0, { error: "network", message: "network error" }) };
  }
}
