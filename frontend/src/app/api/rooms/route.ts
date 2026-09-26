import { bodyOf, library, respond, tokenOf } from "@/lib/libraryGrpc";

export const dynamic = "force-dynamic";

export function POST(req: Request) {
  return respond(bodyOf(req).then((b) => library.createRoom(b, tokenOf(req))), 201);
}
