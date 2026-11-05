"use client";

import { useEffect, useState } from "react";
import type { FormEvent } from "react";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { PenalizedUser, Reservation, Review } from "@/lib/types";
import Button from "./Button";
import AdminLibraries from "./AdminLibraries";
import ErrorState from "./ErrorState";
import Skeleton from "./Skeleton";

function RowSkeleton() {
  return (
    <div className="flex items-center justify-between gap-3 px-4 py-3">
      <Skeleton className="h-4 w-40" />
      <Skeleton className="h-9 w-32" />
    </div>
  );
}

export default function AdminDashboard() {
  const { t, lang } = useT();

  const [pending, setPending] = useState<Reservation[] | null>(null);
  const [pendingError, setPendingError] = useState(false);
  const [pendingBusy, setPendingBusy] = useState<string | null>(null);

  const [reference, setReference] = useState("");
  const [checkinBusy, setCheckinBusy] = useState(false);
  const [checkedIn, setCheckedIn] = useState<Reservation | null>(null);
  const [checkinError, setCheckinError] = useState<string | null>(null);

  const [flagged, setFlagged] = useState<Review[] | null>(null);
  const [flaggedError, setFlaggedError] = useState(false);
  const [reviewBusy, setReviewBusy] = useState<string | null>(null);

  const [penalized, setPenalized] = useState<PenalizedUser[] | null>(null);
  const [penalizedError, setPenalizedError] = useState(false);
  const [unbanBusy, setUnbanBusy] = useState<string | null>(null);

  function loadPenalized() {
    api<PenalizedUser[]>("/api/users/penalized")
      .then((data) => {
        setPenalized(data);
        setPenalizedError(false);
      })
      .catch(() => setPenalizedError(true));
  }

  async function unban(id: string) {
    setUnbanBusy(id);
    try {
      await api(`/api/users/${id}/unban`, { method: "POST" });
      loadPenalized();
    } catch {
      setPenalizedError(true);
    } finally {
      setUnbanBusy(null);
    }
  }

  function loadPending() {
    api<Reservation[]>("/api/reservations?status=pending")
      .then((data) => {
        setPending(data);
        setPendingError(false);
      })
      .catch(() => setPendingError(true));
  }

  function loadFlagged() {
    api<Review[]>("/api/reviews/flagged")
      .then((data) => {
        setFlagged(data);
        setFlaggedError(false);
      })
      .catch(() => setFlaggedError(true));
  }

  useEffect(() => {
    loadPending();
    loadFlagged();
    loadPenalized();
  }, []);

  async function act(id: string, action: "accept" | "reject") {
    setPendingBusy(id);
    try {
      await api(`/api/reservations/${id}/${action}`, { method: "POST" });
      loadPending();
    } catch {
      setPendingError(true);
    } finally {
      setPendingBusy(null);
    }
  }

  async function checkIn(e: FormEvent) {
    e.preventDefault();
    setCheckinBusy(true);
    setCheckedIn(null);
    setCheckinError(null);
    try {
      const res = await api<Reservation>("/api/reservations/checkin", {
        method: "POST",
        body: JSON.stringify({ reference: reference.trim().toUpperCase() }),
      });
      setCheckedIn(res);
      setReference("");
    } catch (err) {
      setCheckinError(t(err instanceof ApiError && err.status === 409 ? "admin.checkinFailed" : "error.generic"));
    } finally {
      setCheckinBusy(false);
    }
  }

  async function moderate(id: string, hidden: boolean) {
    setReviewBusy(id);
    try {
      await api(`/api/reviews/${id}/moderate`, { method: "PATCH", body: JSON.stringify({ hidden }) });
      loadFlagged();
    } catch {
      setFlaggedError(true);
    } finally {
      setReviewBusy(null);
    }
  }

  return (
    <div className="flex flex-col gap-8">
      <h1 className="text-xl font-bold">{t("admin.title")}</h1>

      <section>
        <h2 className="mb-2 text-sm font-semibold">
          {t("admin.pending")}
          {pending && pending.length > 0 ? ` (${pending.length})` : ""}
        </h2>
        {pendingError && <ErrorState onRetry={loadPending} />}
        {pending?.length === 0 && <p className="text-sm text-muted-foreground">{t("admin.noPending")}</p>}
        {pending === null && !pendingError ? (
          <div className="divide-y divide-border rounded-xl border border-border bg-surface">
            <RowSkeleton />
            <RowSkeleton />
          </div>
        ) : (
          pending && pending.length > 0 && (
            <div className="divide-y divide-border rounded-xl border border-border bg-surface">
              {pending.map((r) => (
                <div key={r.id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
                  <div className="text-sm">
                    <p className="font-medium">
                      {new Date(r.start_time).toLocaleString(lang === "th" ? "th-TH" : "en-US")}
                    </p>
                    <p className="text-xs text-muted-foreground">{t("admin.room", { id: r.room_id })}</p>
                  </div>
                  <div className="flex gap-2">
                    <Button variant="secondary" loading={pendingBusy === r.id} onClick={() => act(r.id, "reject")}>
                      {t("admin.reject")}
                    </Button>
                    <Button loading={pendingBusy === r.id} onClick={() => act(r.id, "accept")}>
                      {t("admin.accept")}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )
        )}
      </section>

      <section>
        <h2 className="mb-2 text-sm font-semibold">{t("admin.checkin")}</h2>
        <form onSubmit={checkIn} className="flex flex-wrap items-end gap-2 rounded-xl border border-border bg-surface p-4">
          <label className="flex min-w-48 flex-1 flex-col gap-1 text-xs text-muted-foreground">
            {t("admin.checkinHint")}
            <input
              required
              value={reference}
              onChange={(e) => setReference(e.target.value)}
              placeholder="CU-7F3K9Q"
              autoComplete="off"
              className="min-h-11 rounded-lg border border-border bg-surface px-3 py-2 font-mono text-sm text-foreground outline-none focus-visible:border-brand"
            />
          </label>
          <Button type="submit" loading={checkinBusy}>
            {t("admin.checkinSubmit")}
          </Button>
        </form>
        {checkedIn && (
          <p role="status" className="mt-2 text-sm font-medium text-density-low-fg">
            {t("admin.checkinDone", { reference: checkedIn.reference ?? "", id: checkedIn.room_id })}
          </p>
        )}
        {checkinError && (
          <p role="alert" className="mt-2 text-sm text-red-600 dark:text-red-400">
            {checkinError}
          </p>
        )}
      </section>

      <section>
        <h2 className="mb-2 text-sm font-semibold">
          {t("admin.penalized")}
          {penalized && penalized.length > 0 ? ` (${penalized.length})` : ""}
        </h2>
        {penalizedError && <ErrorState onRetry={loadPenalized} />}
        {penalized?.length === 0 && <p className="text-sm text-muted-foreground">{t("admin.noPenalized")}</p>}
        {penalized && penalized.length > 0 && (
          <div className="divide-y divide-border rounded-xl border border-border bg-surface">
            {penalized.map((u) => (
              <div key={u.id} className="flex flex-wrap items-center justify-between gap-3 px-4 py-3">
                <div className="text-sm">
                  <p className="font-medium">
                    {u.name} · {u.student_id}
                  </p>
                  <p className="text-xs text-muted-foreground">
                    {t("admin.penaltyUntil", {
                      date: new Date(u.expires_at).toLocaleString(lang === "th" ? "th-TH" : "en-US"),
                    })}
                  </p>
                </div>
                <Button variant="secondary" loading={unbanBusy === u.id} onClick={() => unban(u.id)}>
                  {t("admin.unban")}
                </Button>
              </div>
            ))}
          </div>
        )}
      </section>

      <AdminLibraries />

      <section>
        <h2 className="mb-2 text-sm font-semibold">
          {t("admin.reviews")}
          {flagged && flagged.length > 0 ? ` (${flagged.length})` : ""}
        </h2>
        {flaggedError && <ErrorState onRetry={loadFlagged} />}
        {flagged?.length === 0 && <p className="text-sm text-muted-foreground">{t("admin.noFlagged")}</p>}
        {flagged === null && !flaggedError ? (
          <div className="divide-y divide-border rounded-xl border border-border bg-surface">
            <RowSkeleton />
            <RowSkeleton />
          </div>
        ) : (
          flagged &&
          flagged.length > 0 && (
            <div className="divide-y divide-border rounded-xl border border-border bg-surface">
              {flagged.map((rv) => (
                <div key={rv.id} className="flex flex-wrap items-start justify-between gap-3 px-4 py-3">
                  <div className="min-w-0">
                    <p className="text-sm font-medium">
                      {rv.user_name} · {rv.rating}★{" "}
                      {rv.hidden && <span className="ml-1 text-xs text-muted-foreground">({t("admin.hidden")})</span>}
                    </p>
                    {rv.comment && <p className="mt-1 text-sm text-muted-foreground">{rv.comment}</p>}
                  </div>
                  <Button
                    variant={rv.hidden ? "secondary" : "danger"}
                    loading={reviewBusy === rv.id}
                    onClick={() => moderate(rv.id, !rv.hidden)}
                  >
                    {t(rv.hidden ? "admin.unhide" : "admin.hide")}
                  </Button>
                </div>
              ))}
            </div>
          )
        )}
      </section>
    </div>
  );
}
