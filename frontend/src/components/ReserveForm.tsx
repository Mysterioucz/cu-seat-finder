"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { Reservation, Room } from "@/lib/types";
import Button from "./Button";
import Card from "./Card";
import ErrorState from "./ErrorState";
import Skeleton from "./Skeleton";

const DURATIONS = [30, 60, 90, 120] as const;
const fieldClass =
  "min-h-11 rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none focus-visible:border-brand";

// Current local time as datetime-local wants it: the input has no time zone,
// so a UTC string would let Thai users pick up to 7 hours in the past.
function localDateTimeNow() {
  const now = new Date();
  return new Date(now.getTime() - now.getTimezoneOffset() * 60e3).toISOString().slice(0, 16);
}

export default function ReserveForm({ roomId }: { roomId: string }) {
  const { t, lang } = useT();

  const [room, setRoom] = useState<Room | null>(null);
  const [loadError, setLoadError] = useState(false);

  const [start, setStart] = useState("");
  const [duration, setDuration] = useState<(typeof DURATIONS)[number]>(60);
  const [step, setStep] = useState<"form" | "confirm" | "done">("form");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [conflict, setConflict] = useState(false);
  const [reservation, setReservation] = useState<Reservation | null>(null);

  useEffect(() => {
    api<Room>(`/api/rooms/${roomId}`)
      .then(setRoom)
      .catch(() => setLoadError(true));
  }, [roomId]);

  async function confirm() {
    setSubmitting(true);
    setError(null);
    setConflict(false);
    try {
      const res = await api<Reservation>("/api/reservations", {
        method: "POST",
        body: JSON.stringify({
          room_id: roomId,
          start_time: new Date(start).toISOString(),
          duration_minutes: duration,
        }),
      });
      setReservation(res);
      setStep("done");
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        setError(t("reserve.conflict"));
        setConflict(true);
      } else if (err instanceof ApiError && err.code === "quota_exceeded") {
        setError(t("reserve.quota"));
      } else if (err instanceof ApiError && err.code === "penalty") {
        setError(t("reserve.penalty"));
      } else if (err instanceof ApiError && err.code === "forbidden") {
        setError(t("reserve.customerOnly"));
      } else if (err instanceof ApiError && err.code === "start_in_past") {
        setError(t("reserve.startInPast"));
      } else {
        setError(err instanceof ApiError ? err.message : t("error.network"));
      }
    } finally {
      setSubmitting(false);
    }
  }

  if (loadError) return <ErrorState message={t("reserve.notFound")} />;

  return (
    <div className="mx-auto flex max-w-sm flex-col gap-4">
      <h1 className="text-xl font-bold">{t("reserve.title")}</h1>
      <Card>
        {room ? (
          <>
            <p className="font-medium">{room.name}</p>
            <p className="text-xs text-muted-foreground">{room.amenities.join(" · ")}</p>
          </>
        ) : (
          <>
            <Skeleton className="h-5 w-40" />
            <Skeleton className="mt-2 h-3 w-24" />
          </>
        )}
      </Card>

      {step === "form" && (
        <form
          onSubmit={(e) => {
            e.preventDefault();
            setStep("confirm");
          }}
          className="flex flex-col gap-3"
        >
          <label className="flex flex-col gap-1 text-sm">
            {t("reserve.startTime")}
            <input
              required
              type="datetime-local"
              value={start}
              onChange={(e) => setStart(e.target.value)}
              min={localDateTimeNow()}
              className={fieldClass}
            />
          </label>
          <label className="flex flex-col gap-1 text-sm">
            {t("reserve.duration")}
            <select
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value) as (typeof DURATIONS)[number])}
              className={fieldClass}
            >
              {DURATIONS.map((m) => (
                <option key={m} value={m}>
                  {t(`reserve.min${m}` as const)}
                </option>
              ))}
            </select>
          </label>
          <Button type="submit" disabled={!room}>
            {t("reserve.review")}
          </Button>
        </form>
      )}

      {step === "confirm" && (
        <Card className="flex flex-col gap-3">
          <dl className="text-sm">
            <div className="flex justify-between py-1">
              <dt className="text-muted-foreground">{t("reserve.startTime")}</dt>
              <dd>{new Date(start).toLocaleString(lang === "th" ? "th-TH" : "en-US")}</dd>
            </div>
            <div className="flex justify-between py-1">
              <dt className="text-muted-foreground">{t("reserve.duration")}</dt>
              <dd>{t(`reserve.min${duration}` as const)}</dd>
            </div>
          </dl>
          {error && (
            <div role="alert" className="text-sm text-red-600 dark:text-red-400">
              {error}
              {conflict && room && (
                <>
                  {" "}
                  <Link href={`/libraries/${room.library_id}`} className="font-medium text-brand">
                    {t("library.rooms")}
                  </Link>
                </>
              )}
            </div>
          )}
          <div className="flex gap-2">
            <Button variant="secondary" onClick={() => setStep("form")}>
              {t("reserve.back")}
            </Button>
            <Button onClick={confirm} loading={submitting}>
              {submitting ? t("reserve.confirming") : t("reserve.confirm")}
            </Button>
          </div>
        </Card>
      )}

      {step === "done" && reservation && (
        <Card className="flex flex-col gap-2">
          <p className="font-medium text-density-low-fg">{t("reserve.success")}</p>
          <Link href="/me/reservations" className="text-sm font-medium text-brand">
            {t("reserve.viewMyReservations")} →
          </Link>
        </Card>
      )}
    </div>
  );
}
