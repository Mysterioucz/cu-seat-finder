"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { Review } from "@/lib/types";
import Button from "./Button";
import Card from "./Card";
import ReviewForm from "./ReviewForm";
import { StarRatingDisplay } from "./StarRating";

// Owner sees Edit + Delete, admin sees Delete on any review (backend enforces the same rules: 403 otherwise).
export default function ReviewItem({ review: r, userId, isAdmin }: { review: Review; userId?: string; isAdmin: boolean }) {
  const { t } = useT();
  const router = useRouter();
  const [mode, setMode] = useState<"view" | "edit" | "confirm">("view");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const isOwner = !!userId && r.user_id === userId;

  async function remove() {
    setBusy(true);
    setError(null);
    try {
      await api(`/api/reviews/${r.id}`, { method: "DELETE" });
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : t("error.network"));
      setBusy(false);
    }
  }

  if (mode === "edit") {
    return (
      <Card>
        <ReviewForm libraryId={r.library_id} loggedIn review={r} onDone={() => setMode("view")} onCancel={() => setMode("view")} />
      </Card>
    );
  }

  return (
    <Card>
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{r.user_name}</span>
        <StarRatingDisplay value={r.rating} />
      </div>
      {r.tags.length > 0 && (
        <p className="mt-1 text-xs text-muted-foreground">
          {r.tags.map((tag) => t(`tag.${tag}` as Parameters<typeof t>[0])).join(" · ")}
        </p>
      )}
      {r.comment && <p className="mt-2 text-sm">{r.comment}</p>}
      {(isOwner || isAdmin) && (
        <div className="mt-3">
          {mode === "confirm" ? (
            <div role="group" aria-label={t("review.deleteConfirm")} className="flex flex-wrap items-center gap-2">
              <p className="text-sm">{t("review.deleteConfirm")}</p>
              <Button variant="secondary" disabled={busy} onClick={() => setMode("view")} autoFocus>
                {t("action.keep")}
              </Button>
              <Button variant="danger" loading={busy} onClick={remove}>
                {t("review.deleteYes")}
              </Button>
            </div>
          ) : (
            <div className="flex gap-2">
              {isOwner && (
                <Button variant="secondary" onClick={() => setMode("edit")}>
                  {t("action.edit")}
                </Button>
              )}
              <Button variant="secondary" onClick={() => setMode("confirm")}>
                {t("action.delete")}
              </Button>
            </div>
          )}
          {error && (
            <p role="alert" className="mt-2 text-sm text-red-600 dark:text-red-400">
              {error}
            </p>
          )}
        </div>
      )}
    </Card>
  );
}
