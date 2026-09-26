import type { Role } from "./types";

// Non-httpOnly on purpose: client-side fetches attach it as `Authorization: Bearer <token>`
// (the backend reads that header, not a cookie), so browser JS must be able to read it.
// Assumption: acceptable for this project's scope; a production build would put the
// backend behind a same-origin session cookie instead.
export const TOKEN_COOKIE = "cusf_token";

export type JwtClaims = { sub: string; role: Role; name: string; exp: number };

export function decodeToken(token: string): JwtClaims | null {
  try {
    const payload = token.split(".")[1];
    const json = atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    const claims = JSON.parse(json) as JwtClaims;
    if (!claims.exp || claims.exp * 1000 < Date.now()) return null;
    return claims;
  } catch {
    return null;
  }
}

export function readCookie(name: string): string | undefined {
  if (typeof document === "undefined") return undefined;
  const match = document.cookie.match(new RegExp(`(?:^|; )${name}=([^;]*)`));
  return match ? decodeURIComponent(match[1]) : undefined;
}

export function setToken(token: string) {
  document.cookie = `${TOKEN_COOKIE}=${encodeURIComponent(token)}; path=/; max-age=${60 * 60 * 12}; samesite=lax`;
}

export function clearToken() {
  document.cookie = `${TOKEN_COOKIE}=; path=/; max-age=0`;
}

export function getClientUser(): JwtClaims | null {
  const token = readCookie(TOKEN_COOKIE);
  return token ? decodeToken(token) : null;
}
