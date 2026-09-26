import { bodyOf, library, respond, tokenOf } from "@/lib/libraryGrpc";

export const dynamic = "force-dynamic";
type Ctx = { params: Promise<{ id: string }> };

export async function GET(_: Request, { params }: Ctx) {
  return respond(library.get((await params).id));
}

export async function PUT(req: Request, { params }: Ctx) {
  const { id } = await params;
  return respond(bodyOf(req).then((b) => library.update({ ...b, id }, tokenOf(req))));
}

export async function DELETE(req: Request, { params }: Ctx) {
  return respond(library.remove((await params).id, tokenOf(req)));
}
