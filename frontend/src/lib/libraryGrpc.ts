// Library Service is gRPC-only (:9082, proto/library/v1/library.proto). This is the only place the
// Next.js server (the BFF) talks to it: route handlers under app/api/{libraries,rooms} expose REST,
// server components call `library.*` directly (no self-HTTP hop).
import "server-only";
import path from "node:path";
import { NextResponse } from "next/server";
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import { ApiError } from "./apiError";
import type { Library, Room } from "./types";

const ADDR = process.env.LIBRARY_GRPC_ADDR ?? "localhost:9082";
// process.cwd() is frontend/ for both `next dev` and `next build/start`.
const PROTO = path.join(process.cwd(), "..", "proto", "library", "v1", "library.proto");

// gRPC status -> HTTP per API-CONTRACT.md
const HTTP: Record<number, number> = { 3: 400, 5: 404, 7: 403, 9: 409, 14: 503, 16: 401 };

type Client = grpc.Client & Record<string, (req: unknown, md: grpc.Metadata, opts: grpc.CallOptions, cb: (e: grpc.ServiceError | null, res: unknown) => void) => void>;
const g = globalThis as { __libraryClient?: Client };

function client(): Client {
  if (!g.__libraryClient) {
    const def = protoLoader.loadSync(PROTO, { keepCase: true, longs: Number, enums: String, defaults: true, arrays: true });
    const pkg = grpc.loadPackageDefinition(def) as unknown as { library: { v1: { LibraryService: typeof grpc.Client } } };
    g.__libraryClient = new pkg.library.v1.LibraryService(ADDR, grpc.credentials.createInsecure()) as Client;
  }
  return g.__libraryClient;
}

function rpc<T>(method: string, req: object, token?: string | null): Promise<T> {
  return new Promise((resolve, reject) => {
    const md = new grpc.Metadata();
    if (token) md.set("authorization", token.startsWith("Bearer ") ? token : `Bearer ${token}`);
    client()[method](req, md, { deadline: Date.now() + 5000 }, (err, res) => {
      if (!err) return resolve(res as T);
      const status = HTTP[err.code] ?? 500;
      reject(new ApiError(status, { error: (grpc.status[err.code] ?? "internal").toLowerCase(), message: err.details || err.message }));
    });
  });
}

// Keep the old JSON shape: empty image_url is omitted.
const clean = (l: Library): Library => {
  if (!l.image_url) delete l.image_url;
  return l;
};

export const library = {
  list: (q = "", amenity = "") =>
    rpc<{ libraries: Library[] }>("ListLibraries", { q, amenity }).then((r) => r.libraries.map(clean)),
  get: (id: string) => rpc<Library>("GetLibrary", { id }).then(clean),
  create: (l: Partial<Library>, token?: string | null) => rpc<Library>("CreateLibrary", { library: l }, token).then(clean),
  update: (l: Partial<Library>, token?: string | null) => rpc<Library>("UpdateLibrary", { library: l }, token).then(clean),
  remove: (id: string, token?: string | null) => rpc("DeleteLibrary", { id }, token).then(() => undefined),
  rooms: (libraryId = "") => rpc<{ rooms: Room[] }>("ListRooms", { library_id: libraryId }).then((r) => r.rooms),
  room: (id: string) => rpc<Room>("GetRoom", { id }),
  createRoom: (room: Partial<Room>, token?: string | null) => rpc<Room>("CreateRoom", { room }, token),
  updateRoom: (room: Partial<Room>, token?: string | null) => rpc<Room>("UpdateRoom", { room }, token),
  removeRoom: (id: string, token?: string | null) => rpc("DeleteRoom", { id }, token).then(() => undefined),
};

/** Server components: never throws, same {data,error} shape as serverApi. */
export async function safe<T>(p: Promise<T>): Promise<{ data: T | null; error: ApiError | null }> {
  try {
    return { data: await p, error: null };
  } catch (e) {
    return { data: null, error: e instanceof ApiError ? e : new ApiError(0, { error: "network", message: "network error" }) };
  }
}

/** Route handlers: JSON body on success (204 for void), {error,message} + mapped status on failure. */
export async function respond<T>(p: Promise<T>, ok = 200): Promise<NextResponse> {
  const nocache = { headers: { "Cache-Control": "no-store" } };
  try {
    const data = await p;
    return data === undefined ? new NextResponse(null, { status: 204, ...nocache }) : NextResponse.json(data, { status: ok, ...nocache });
  } catch (e) {
    const err = e instanceof ApiError ? e : new ApiError(500, { error: "internal", message: "internal error" });
    return NextResponse.json({ error: err.code, message: err.message }, { status: err.status, ...nocache });
  }
}

/** Parse JSON body; rejects with a 400 ApiError on malformed input. */
export const bodyOf = (req: Request) =>
  req.json().catch(() => {
    throw new ApiError(400, { error: "invalid_argument", message: "invalid JSON body" });
  }) as Promise<Record<string, unknown>>;
export const tokenOf = (req: Request) => req.headers.get("authorization");
