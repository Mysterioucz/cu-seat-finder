import type { HTMLAttributes } from "react";

export default function Card({ className = "", ...rest }: HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={`rounded-xl border border-border bg-surface p-4 shadow-sm dark:shadow-none ${className}`}
      {...rest}
    />
  );
}
