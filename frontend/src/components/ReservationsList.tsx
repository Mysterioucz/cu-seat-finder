"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { api } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { Reservation, ReservationStatus } from "@/lib/types";
import Button from "./Button";
import Card from "./Card";
import Countdown from "./Countdown";
import ErrorState from "./ErrorState";
import Skeleton from "./Skeleton";

const STATUS_STYLE: Record<ReservationStatus, string> = {
  pending: "bg-density-medium-bg text-density-medium-fg",
  accepted: "bg-density-low-bg text-density-low-fg",
  rejected: "bg-density-full-bg text-density-full-fg",
  cancelled: "bg-surface-muted text-muted-foreground",
  no_show: "bg-density-full-bg text-density-full-fg",
};

function CardSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <div className="flex items-start justify-between gap-3">
        <Skeleton className="h-4 w-40" />
        <Skeleton className="h-6 w-16 rounded-full" />
      </div>
      <Skeleton className="mt-3 h-16 w-full" />
    </div>
  );
}

export default function ReservationsList() {
  const { t, lang } = useT();
  const [items, setItems] = useState<Reservation[] | null>(null);
  const [error, setError] = useState(false);
  const [cancellingId, setCancellingId] = useState<string | null>(null);

  function load() {
    api<Reservation[]>("/api/reservations/me")
      .then((data) => {
        setItems(data);
        setError(false);
      })
      .catch(() => setError(true));
  }

  useEffect(load, []);

  async function cancel(id: string) {
    setCancellingId(id);
    try {
      await api(`/api/reservations/${id}/cancel`, { method: "POST" });
      load();
    } catch {
      setError(true);
    } finally {
      setCancellingId(null);
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-bold">{t("reservations.title")}</h1>
      {error && <ErrorState onRetry={load} />}

      {items === null && !error && (
        <div className="flex flex-col gap-3">
          <CardSkeleton />
          <CardSkeleton />
        </div>
      )}

      {items?.length === 0 && (
        <Card className="flex flex-col items-start gap-2">
          <p className="text-sm text-muted-foreground">{t("reservations.empty")}</p>
          <Link href="/" className="text-sm font-medium text-brand">
            {t("reservations.browse")}
          </Link>
        </Card>
      )}

      <div className="flex flex-col gap-3">
        {items?.map((r) => {
          const isTicket = r.status === "accepted" && r.reference && r.checkin_deadline;
          return (
            <Card key={r.id} className={isTicket ? "overflow-hidden p-0" : undefined}>
              <div className={`flex items-start justify-between gap-3 ${isTicket ? "p-4" : ""}`}>
                <div>
                  <p className="text-sm font-medium">
                    {new Date(r.start_time).toLocaleString(lang === "th" ? "th-TH" : "en-US")}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {new Date(r.end_time).toLocaleTimeString(lang === "th" ? "th-TH" : "en-US")}
                  </p>
                </div>
                <span className={`rounded-full px-2.5 py-1 text-xs font-semibold ${STATUS_STYLE[r.status]}`}>
                  {t(`reservations.status.${r.status}`)}
                </span>
              </div>

              {isTicket && r.reference && r.checkin_deadline && (
                // Ticket-stub styling: dashed tear line + shaded stub, so the check-in
                // credential reads as something you'd hand over, not just another row.
                <div className="flex items-center gap-4 border-t border-dashed border-border bg-surface-muted p-4">
                  <div className="shrink-0 rounded-lg bg-white p-1.5">
                    <QRCodeSVG value={r.reference} size={72} title={r.reference} />
                  </div>
                  <div className="min-w-0 text-sm">
                    <p className="text-xs text-muted-foreground">{t("reservations.reference")}</p>
                    <p className="font-mono text-lg font-semibold tracking-wide">{r.reference}</p>
                    <p className="mt-2 text-xs text-muted-foreground">{t("reservations.checkinBy")}</p>
                    {/* Plain foreground, not --brand: brand text on the tinted stub bg
                        falls under 4.5:1 in light mode. Size + weight carry the emphasis. */}
                    <p className="text-lg font-semibold">
                      <Countdown deadline={r.checkin_deadline} expiredLabel={t("reservations.expired")} />
                    </p>
                  </div>
                </div>
              )}

              {(r.status === "pending" || r.status === "accepted") && (
                <div className={isTicket ? "p-4 pt-0" : "mt-3"}>
                  <Button
                    variant="secondary"
                    loading={cancellingId === r.id}
                    onClick={() => cancel(r.id)}
                  >
                    {t("reservations.cancel")}
                  </Button>
                </div>
              )}
            </Card>
          );
        })}
      </div>
    </div>
  );
}
