import { bodyOf, library, respond, tokenOf } from "@/lib/libraryGrpc";

export const dynamic = "force-dynamic";

export function GET(req: Request) {
  const sp = new URL(req.url).searchParams;
  return respond(library.list(sp.get("q") ?? "", sp.get("amenity") ?? ""));
}

export function POST(req: Request) {
  return respond(bodyOf(req).then((b) => library.create(b, tokenOf(req))), 201);
}
