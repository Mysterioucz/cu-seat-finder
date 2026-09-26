"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";
import { api } from "@/lib/api";
import { useT } from "@/lib/LangContext";
import type { Lang } from "@/lib/i18n";
import type { Library, Occupancy, OccupancyResponse } from "@/lib/types";
import Button from "./Button";
import Chip from "./Chip";
import ErrorState from "./ErrorState";
import OccupancyBadge, { OccupancyBar } from "./OccupancyBadge";
import Skeleton from "./Skeleton";

const POLL_MS = 30_000;

function minutesAgo(iso: string): number {
  return Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 60_000));
}

function LibraryRow({ library, occupancy, lang }: { library: Library; occupancy: Occupancy; lang: Lang }) {
  const { t } = useT();
  const name = lang === "th" ? library.name_th : library.name_en;
  return (
    <Link
      href={`/libraries/${library.id}`}
      className="block rounded-xl border border-border bg-surface p-4 shadow-sm transition-colors hover:border-brand dark:shadow-none"
    >
      <div className="flex items-start justify-between gap-3">
        <h3 className="font-semibold">{name}</h3>
        <OccupancyBadge percent={occupancy.percent} />
      </div>
      <div className="mt-3">
        <OccupancyBar occupancy={occupancy} />
      </div>
      {occupancy.stale && (
        <p className="mt-2 text-xs text-density-medium-fg" role="status">
          {t("home.lastUpdated", { n: minutesAgo(occupancy.updated_at) })}
        </p>
      )}
    </Link>
  );
}

function RowSkeleton() {
  return (
    <div className="rounded-xl border border-border bg-surface p-4">
      <div className="flex items-start justify-between gap-3">
        <Skeleton className="h-5 w-32" />
        <Skeleton className="h-6 w-20 rounded-full" />
      </div>
      <Skeleton className="mt-3 h-2 w-full" />
      <Skeleton className="mt-2 h-3 w-24" />
    </div>
  );
}

export default function LibrariesList({
  lang,
  initialLibraries,
  initialOccupancy,
  initialError,
}: {
  lang: Lang;
  initialLibraries: Library[] | null;
  initialOccupancy: OccupancyResponse | null;
  initialError: boolean;
}) {
  const { t } = useT();
  const [libraries, setLibraries] = useState<Library[] | null>(initialLibraries);
  const [occupancy, setOccupancy] = useState<OccupancyResponse | null>(initialOccupancy);
  const [libError, setLibError] = useState(initialError);
  const [occError, setOccError] = useState(initialError);
  const [min, setMin] = useState(0);
  const [max, setMax] = useState(100);
  const [amenities, setAmenities] = useState<string[]>([]);

  useEffect(() => {
    if (libraries) return;
    api<Library[]>("/api/libraries")
      .then((data) => {
        setLibraries(data);
        setLibError(false);
      })
      .catch(() => setLibError(true));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    let cancelled = false;
    function poll() {
      api<OccupancyResponse>(`/api/occupancy?min=${min}&max=${max}`)
        .then((data) => {
          if (cancelled) return;
          setOccupancy(data);
          setOccError(false);
        })
        .catch(() => {
          if (!cancelled) setOccError(true);
        });
    }
    poll();
    const id = setInterval(poll, POLL_MS);
    return () => {
      cancelled = true;
      clearInterval(id);
    };
  }, [min, max]);

  const allAmenities = useMemo(
    () => Array.from(new Set((libraries ?? []).flatMap((l) => l.amenities))).sort(),
    [libraries],
  );

  const libraryById = useMemo(() => new Map((libraries ?? []).map((l) => [l.id, l])), [libraries]);

  const rows = useMemo(() => {
    if (!occupancy) return [];
    return occupancy.items
      .map((occ) => ({ occ, lib: libraryById.get(occ.library_id) }))
      .filter((r): r is { occ: Occupancy; lib: Library } => !!r.lib)
      .filter((r) => amenities.every((a) => r.lib.amenities.includes(a)));
  }, [occupancy, libraryById, amenities]);

  const suggestionLib = occupancy?.suggestion ? libraryById.get(occupancy.suggestion.library_id) : undefined;

  const activeFilterCount = (min > 0 ? 1 : 0) + (max < 100 ? 1 : 0) + amenities.length;

  function toggleAmenity(a: string) {
    setAmenities((prev) => (prev.includes(a) ? prev.filter((x) => x !== a) : [...prev, a]));
  }

  function resetFilters() {
    setMin(0);
    setMax(100);
    setAmenities([]);
  }

  if (libError && occError) {
    return (
      <ErrorState
        message={t("error.network")}
        onRetry={() => {
          setLibError(false);
          setOccError(false);
          setLibraries(null);
        }}
      />
    );
  }

  const loading = !libraries && !libError;

  return (
    <div className="flex flex-col gap-4">
      <h1 className="text-xl font-bold">{t("home.title")}</h1>

      {/* Collapsed by default: filters must never push the answer to "where can I sit"
          below the fold on a phone screen. */}
      <details className="group rounded-xl border border-border bg-surface">
        <summary className="flex min-h-11 cursor-pointer list-none items-center justify-between gap-2 px-4 py-3 text-sm font-semibold [&::-webkit-details-marker]:hidden">
          <span>
            {activeFilterCount > 0 ? t("home.filtersActive", { n: activeFilterCount }) : t("home.filtersToggle")}
          </span>
          <span aria-hidden className="text-muted-foreground transition-transform group-open:rotate-180">
            ▾
          </span>
        </summary>
        <div className="flex flex-col gap-4 border-t border-border px-4 pb-4 pt-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <label className="flex flex-col gap-1 text-sm">
              {t("home.minOccupancy")}: {min}%
              <input
                type="range"
                min={0}
                max={100}
                value={min}
                onChange={(e) => setMin(Math.min(Number(e.target.value), max))}
                className="h-11"
              />
            </label>
            <label className="flex flex-col gap-1 text-sm">
              {t("home.maxOccupancy")}: {max}%
              <input
                type="range"
                min={0}
                max={100}
                value={max}
                onChange={(e) => setMax(Math.max(Number(e.target.value), min))}
                className="h-11"
              />
            </label>
          </div>
          {allAmenities.length > 0 && (
            <div>
              <p className="mb-2 text-sm font-medium">{t("home.amenities")}</p>
              <div className="flex flex-wrap gap-2">
                {allAmenities.map((a) => (
                  <Chip key={a} label={a} pressed={amenities.includes(a)} onClick={() => toggleAmenity(a)} />
                ))}
              </div>
            </div>
          )}
          {activeFilterCount > 0 && (
            <Button variant="ghost" className="self-start px-0" onClick={resetFilters}>
              {t("home.resetFilters")}
            </Button>
          )}
        </div>
      </details>

      {occError && !libError && <ErrorState message={t("error.network")} />}

      {loading ? (
        <div className="flex flex-col gap-3">
          <RowSkeleton />
          <RowSkeleton />
          <RowSkeleton />
        </div>
      ) : rows.length === 0 ? (
        <div className="flex flex-col gap-3">
          <p className="text-sm text-muted-foreground">{t("home.noResults")}</p>
          {suggestionLib && occupancy?.suggestion && (
            <>
              <p className="text-sm font-medium">{t("home.suggestion")}</p>
              <LibraryRow library={suggestionLib} occupancy={occupancy.suggestion} lang={lang} />
            </>
          )}
        </div>
      ) : (
        <div className="flex flex-col gap-3">
          <p className="text-xs text-muted-foreground">{t("home.resultCount", { n: rows.length })}</p>
          {rows.map(({ occ, lib }) => (
            <LibraryRow key={lib.id} library={lib} occupancy={occ} lang={lang} />
          ))}
        </div>
      )}
    </div>
  );
}
