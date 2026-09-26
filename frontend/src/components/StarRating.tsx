"use client";

const STAR = "★";

export function StarRatingDisplay({ value, count }: { value: number; count?: number }) {
  return (
    <span className="inline-flex items-center gap-1" aria-label={`${value.toFixed(1)} / 5`}>
      <span className="text-amber-500" aria-hidden>
        {[1, 2, 3, 4, 5].map((i) => (
          <span key={i} className={i <= Math.round(value) ? "" : "opacity-25"}>
            {STAR}
          </span>
        ))}
      </span>
      <span className="text-sm text-muted-foreground">
        {value.toFixed(1)}
        {typeof count === "number" ? ` (${count})` : ""}
      </span>
    </span>
  );
}

export function StarRatingInput({
  value,
  onChange,
  label,
}: {
  value: number;
  onChange: (v: number) => void;
  label: string;
}) {
  return (
    <div role="radiogroup" aria-label={label} className="flex gap-0.5">
      {[1, 2, 3, 4, 5].map((i) => (
        <button
          key={i}
          type="button"
          role="radio"
          aria-checked={value === i}
          aria-label={`${i} / 5`}
          onClick={() => onChange(i)}
          className={`flex min-h-11 min-w-11 items-center justify-center text-2xl leading-none ${
            i <= value ? "text-amber-500" : "text-muted-foreground/40"
          }`}
        >
          {STAR}
        </button>
      ))}
    </div>
  );
}
