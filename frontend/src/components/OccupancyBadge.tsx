"use client";

import { useT } from "@/lib/LangContext";
import type { Occupancy } from "@/lib/types";

type Density = "low" | "medium" | "high" | "full";

// 4 bands, not 3: a 100%-full room reads differently from an 80%-full one when a
// student is deciding whether it's worth walking over.
function densityOf(percent: number): Density {
  if (percent >= 95) return "full";
  if (percent >= 70) return "high";
  if (percent >= 40) return "medium";
  return "low";
}

const styles: Record<Density, string> = {
  low: "bg-density-low-bg text-density-low-fg",
  medium: "bg-density-medium-bg text-density-medium-fg",
  high: "bg-density-high-bg text-density-high-fg",
  full: "bg-density-full-bg text-density-full-fg",
};

const barStyles: Record<Density, string> = {
  low: "bg-density-low-fg",
  medium: "bg-density-medium-fg",
  high: "bg-density-high-fg",
  full: "bg-density-full-fg",
};

const dots = { low: "●", medium: "●●", high: "●●●", full: "●●●●" } as const;

// Never color-only: badge pairs color with a dot glyph + text density label,
// and the bar below always ships with an explicit free-seat count.
export default function OccupancyBadge({ percent }: { percent: number }) {
  const { t } = useT();
  const density = densityOf(percent);
  const label = t(`occupancy.${density}`);
  return (
    <span
      className={`inline-flex shrink-0 items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-semibold ${styles[density]}`}
    >
      <span aria-hidden>{dots[density]}</span>
      {percent}% · {label}
    </span>
  );
}

/** Capacity bar + free-seat count — the actual answer to "where can I sit right now".
 * Renders alongside OccupancyBadge wherever occupancy shows, not instead of it. */
export function OccupancyBar({ occupancy }: { occupancy: Occupancy }) {
  const { t } = useT();
  const density = densityOf(occupancy.percent);
  const free = Math.max(0, occupancy.capacity - occupancy.current);
  return (
    <div className="flex flex-col gap-1">
      <div
        role="progressbar"
        aria-valuenow={occupancy.percent}
        aria-valuemin={0}
        aria-valuemax={100}
        aria-label={t("occupancy.ariaBar")}
        className="h-2 w-full overflow-hidden rounded-full bg-surface-muted"
      >
        <div
          className={`h-full rounded-full ${barStyles[density]}`}
          style={{ width: `${Math.min(100, Math.max(0, occupancy.percent))}%` }}
        />
      </div>
      <p className="text-xs text-muted-foreground">{t("home.freeSeats", { n: free })}</p>
    </div>
  );
}
