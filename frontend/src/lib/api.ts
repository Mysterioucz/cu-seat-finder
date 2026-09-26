// Client-side fetch helper. Always hits our own origin (:3000) — next.config.ts
// rewrites forward /api/* to the right backend service. Adds the bearer token and
// parses the {error,message} error shape from the contract.
import { ApiError } from "./apiError";
import { readCookie, TOKEN_COOKIE } from "./auth";

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const token = readCookie(TOKEN_COOKIE);
  let res: Response;
  try {
    res = await fetch(path, {
      ...init,
      cache: init.cache ?? "no-store",
      headers: {
        ...(init.body ? { "Content-Type": "application/json" } : {}),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...init.headers,
      },
    });
  } catch {
    throw new ApiError(0, { error: "network", message: "network error" });
  }
  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => null);
  if (!res.ok) {
    throw new ApiError(res.status, data ?? { error: "unknown", message: "Request failed" });
  }
  return data as T;
}

export { ApiError };
