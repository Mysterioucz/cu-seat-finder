"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { api, ApiError } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import { REVIEW_TAGS, type Review } from "@/lib/types";
import Button from "./Button";
import Chip from "./Chip";
import { StarRatingInput } from "./StarRating";

// loggedIn is decided server-side (see libraries/[id]/page.tsx) so there's no
// login-state flash: the cookie is readable on the server even though it's also
// readable by client JS (needed so the browser can attach it as a Bearer header).
// Pass `review` (+ onDone/onCancel) to edit an existing review in place: prefilled, PUT instead of POST.
export default function ReviewForm({
  libraryId,
  loggedIn,
  review,
  onDone,
  onCancel,
}: {
  libraryId: string;
  loggedIn: boolean;
  review?: Review;
  onDone?: () => void;
  onCancel?: () => void;
}) {
  const { t } = useT();
  const router = useRouter();
  const [rating, setRating] = useState(review?.rating ?? 0);
  const [tags, setTags] = useState<string[]>(review?.tags ?? []);
  const [comment, setComment] = useState(review?.comment ?? "");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [done, setDone] = useState(false);

  if (!loggedIn) {
    return (
      <p className="text-sm text-muted-foreground">
        {t("review.loginRequired")} <a href="/login" className="font-medium text-brand">{t("nav.login")}</a>
      </p>
    );
  }

  if (done) return <p className="text-sm font-medium text-green-700 dark:text-green-400">{t("review.success")}</p>;

  function toggleTag(tag: string) {
    setTags((prev) => (prev.includes(tag) ? prev.filter((x) => x !== tag) : [...prev, tag]));
  }

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    if (rating < 1) {
      setError(t("review.rating"));
      return;
    }
    setError(null);
    setSubmitting(true);
    try {
      await api(review ? `/api/reviews/${review.id}` : "/api/reviews", {
        method: review ? "PUT" : "POST",
        body: JSON.stringify({ library_id: libraryId, rating, tags, comment: comment || undefined }),
      });
      if (review) onDone?.();
      else setDone(true);
      router.refresh();
    } catch (err) {
      if (err instanceof ApiError && err.code === "profanity") {
        setError(t("review.profanity"));
      } else {
        setError(err instanceof ApiError ? err.message : t("error.network"));
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form onSubmit={submit} className="flex flex-col gap-3">
      <div>
        <p className="mb-1 text-sm font-medium">{t("review.rating")}</p>
        <StarRatingInput value={rating} onChange={setRating} label={t("review.rating")} />
      </div>
      <div>
        <p className="mb-2 text-sm font-medium">{t("review.tags")}</p>
        <div className="flex flex-wrap gap-2">
          {REVIEW_TAGS.map((tag) => (
            <Chip key={tag} label={t(`tag.${tag}`)} pressed={tags.includes(tag)} onClick={() => toggleTag(tag)} />
          ))}
        </div>
      </div>
      <label className="flex flex-col gap-1 text-sm">
        {t("review.comment")}
        <textarea
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          placeholder={t("review.commentPlaceholder")}
          rows={3}
          className="rounded-lg border border-border bg-surface px-3 py-2 text-sm outline-none focus-visible:border-brand"
        />
      </label>
      {error && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
      <div className="flex gap-2">
        {review && (
          <Button type="button" variant="secondary" onClick={onCancel}>
            {t("action.cancel")}
          </Button>
        )}
        <Button type="submit" loading={submitting}>
          {submitting ? t("review.submitting") : t(review ? "action.save" : "review.submit")}
        </Button>
      </div>
    </form>
  );
}
