import { getLang } from "@/lib/langServer";
import { serverApi } from "@/lib/serverApi";
import { library, safe } from "@/lib/libraryGrpc";
import type { OccupancyResponse } from "@/lib/types";
import LibrariesList from "@/components/LibrariesList";

export const dynamic = "force-dynamic";

export default async function HomePage() {
  const lang = await getLang();
  const [libRes, occRes] = await Promise.all([
    safe(library.list()),
    serverApi<OccupancyResponse>("occupancy", "/api/occupancy"),
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
