"use client";

import { useEffect, useState } from "react";

function format(ms: number): string {
  const totalSec = Math.max(0, Math.floor(ms / 1000));
  const m = Math.floor(totalSec / 60);
  const s = totalSec % 60;
  return `${m}:${s.toString().padStart(2, "0")}`;
}

// Ticking clock: server-rendered and first-hydrated output will always be a beat
// behind the client's current time, so the mismatch is deliberate — suppressed below.
export default function Countdown({ deadline, expiredLabel }: { deadline: string; expiredLabel: string }) {
  const target = new Date(deadline).getTime();
  const [now, setNow] = useState(() => Date.now());

  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), 1000);
    return () => clearInterval(id);
  }, []);

  const remaining = target - now;
  if (remaining <= 0) {
    return (
      <span suppressHydrationWarning className="font-medium text-red-600">
        {expiredLabel}
      </span>
    );
  }
  return (
    <span suppressHydrationWarning className="font-mono font-medium">
      {format(remaining)}
    </span>
  );
}
