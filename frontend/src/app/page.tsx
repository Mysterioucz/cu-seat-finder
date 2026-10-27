import { getLang } from "@/lib/langServer";
import { serverApi } from "@/lib/serverApi";
import type { Library, OccupancyResponse } from "@/lib/types";
import LibrariesList from "@/components/LibrariesList";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const lang = await getLang();
  const [libRes, occRes] = await Promise.all([
    serverApi<Library[]>("/api/libraries"),
    serverApi<OccupancyResponse>("/api/occupancy"),
  ]);

  return (
    <LibrariesList
      lang={lang}
      initialLibraries={libRes.data}
      initialOccupancy={occRes.data}
      initialError={!libRes.data && !occRes.data}
    />
  );
}
