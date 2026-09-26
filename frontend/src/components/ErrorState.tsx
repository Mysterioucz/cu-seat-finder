"use client";

import { useT } from "@/lib/LangContext";
import Button from "./Button";

export default function ErrorState({ message, onRetry }: { message?: string; onRetry?: () => void }) {
  const { t } = useT();
  return (
    <div
      role="alert"
      className="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-300"
    >
      <p>{message ?? t("error.generic")}</p>
      {onRetry && (
        <Button variant="danger" className="mt-3" onClick={onRetry}>
          {t("action.retry")}
        </Button>
      )}
    </div>
  );
}
