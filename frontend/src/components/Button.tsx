import type { ButtonHTMLAttributes } from "react";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "danger" | "ghost";
  loading?: boolean;
};

// min-h-11 (44px) on every variant: WCAG 2.5.8 target-size minimum for primary actions,
// and cheaper to apply uniformly than to special-case which buttons count as "primary".
const base =
  "inline-flex min-h-11 items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed";

const variants: Record<NonNullable<Props["variant"]>, string> = {
  primary: "bg-brand text-brand-foreground hover:opacity-90",
  secondary: "bg-surface-muted text-foreground hover:opacity-80 border border-border",
  danger: "bg-red-600 text-white hover:bg-red-700",
  ghost: "bg-transparent text-current hover:bg-surface-muted",
};

export default function Button({ variant = "primary", loading, className = "", children, disabled, ...rest }: Props) {
  return (
    <button
      className={`${base} ${variants[variant]} ${className}`}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...rest}
    >
      {children}
    </button>
  );
}
