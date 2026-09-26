import { library, respond } from "@/lib/libraryGrpc";

export const dynamic = "force-dynamic";

export async function GET(_: Request, { params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  // ListRooms treats empty library_id as "all rooms", so verify the library exists first (404).
  return respond(library.get(id).then(() => library.rooms(id)));
}
