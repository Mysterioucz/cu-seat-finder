import type { HTMLAttributes } from "react";

/** Lightweight loading placeholder. Reused across reservations/admin/reserve so those
 * screens don't flash bare "loading..." text. Respects prefers-reduced-motion globally
 * (see .skeleton in globals.css). */
export default function Skeleton({ className = "", ...rest }: HTMLAttributes<HTMLDivElement>) {
  return <div aria-hidden className={`skeleton ${className}`} {...rest} />;
}
