import { cookies } from "next/headers";
import { decodeToken, TOKEN_COOKIE, type JwtClaims } from "./auth";

/** Reads and decodes the auth JWT server-side. httpOnly doesn't block server reads —
 * only client JS — so pages can gate on role/login without a client round trip. */
export async function getAuthUser(): Promise<JwtClaims | null> {
  const store = await cookies();
  const token = store.get(TOKEN_COOKIE)?.value;
  return token ? decodeToken(token) : null;
}
