import Link from "next/link";
import { getLang } from "@/lib/langServer";
import { getAuthUser } from "@/lib/authServer";
import { serverApi } from "@/lib/serverApi";
import { t } from "@/lib/i18n";
import type { Library, Occupancy, ReviewsResponse, Room } from "@/lib/types";
import Card from "@/components/Card";
import OccupancyBadge, { OccupancyBar } from "@/components/OccupancyBadge";
import { StarRatingDisplay } from "@/components/StarRating";
import ReviewForm from "@/components/ReviewForm";
import ReviewItem from "@/components/ReviewItem";

export const dynamic = "force-dynamic";

export default async function LibraryDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const lang = await getLang();
  const user = await getAuthUser();

  const [libRes, roomsRes, reviewsRes, occRes] = await Promise.all([
    serverApi<Library>(`/api/libraries/${id}`),
    serverApi<Room[]>(`/api/libraries/${id}/rooms`),
    serverApi<ReviewsResponse>(`/api/reviews?library_id=${id}`),
    serverApi<Occupancy>(`/api/occupancy/${id}`),
  ]);

  if (!libRes.data) {
    const message = libRes.error?.status === 0 || libRes.error?.status === 503 ? t(lang, "error.network") : t(lang, "library.notFound");
    return <p className="text-sm text-muted-foreground">{message}</p>;
  }

  const library = libRes.data;
  const rooms = roomsRes.data ?? [];
  const reviews = reviewsRes.data ?? { average: 0, count: 0, items: [] };
  const name = lang === "th" ? library.name_th : library.name_en;

  return (
    <div className="flex flex-col gap-6">
      <div>
        <div className="flex items-start justify-between gap-3">
          <div>
            <h1 className="text-xl font-bold">{name}</h1>
            <p className="text-sm text-muted-foreground">{library.amenities.join(" · ")}</p>
          </div>
          {occRes.data && <OccupancyBadge percent={occRes.data.percent} />}
        </div>
        {occRes.data && (
          <div className="mt-3 max-w-sm">
            <OccupancyBar occupancy={occRes.data} />
          </div>
        )}
      </div>

      <section>
        <h2 className="mb-2 text-sm font-semibold">{t(lang, "library.floors")}</h2>
        <div className="flex flex-col gap-2">
          {library.floors.map((floor) => (
            <Card key={floor.id}>
              <p className="font-medium">{floor.name}</p>
              <ul className="mt-1 flex flex-wrap gap-x-4 gap-y-1 text-sm text-muted-foreground">
                {floor.zones.map((z) => (
                  <li key={z.id}>
                    {z.name} ({z.capacity})
                  </li>
                ))}
              </ul>
            </Card>
          ))}
        </div>
      </section>

      <section>
        <h2 className="mb-2 text-sm font-semibold">{t(lang, "library.rooms")}</h2>
        <div className="flex flex-col gap-2">
          {rooms.map((room) => (
            <Card key={room.id} className="flex items-center justify-between gap-3">
              <div>
                <p className="font-medium">{room.name}</p>
                <p className="text-xs text-muted-foreground">{room.amenities.join(" · ")}</p>
              </div>
              <Link
                href={`/reserve/${room.id}`}
                className="inline-flex min-h-11 shrink-0 items-center rounded-lg bg-brand px-4 text-sm font-medium text-brand-foreground hover:opacity-90"
              >
                {t(lang, "library.book")}
              </Link>
            </Card>
          ))}
        </div>
      </section>

      <section>
        <div className="mb-2 flex items-center justify-between">
          <h2 className="text-sm font-semibold">{t(lang, "library.reviews")}</h2>
          <StarRatingDisplay value={reviews.average} count={reviews.count} />
        </div>
        <div className="mb-4 flex flex-col gap-2">
          {reviews.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">{t(lang, "library.noReviews")}</p>
          ) : (
            reviews.items.map((r) => (
              <ReviewItem key={r.id} review={r} userId={user?.sub} isAdmin={user?.role === "admin"} />
            ))
          )}
        </div>
        <Card>
          <h3 className="mb-3 text-sm font-semibold">{t(lang, "library.writeReview")}</h3>
          <ReviewForm libraryId={library.id} loggedIn={!!user} />
        </Card>
      </section>
    </div>
  );
}
