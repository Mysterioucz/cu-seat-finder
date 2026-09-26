// Base URLs for the 4 REST backend services (Library is gRPC: see libraryGrpc.ts) (see .planning/API-CONTRACT.md topology table).
// Env-overridable. Used by next.config.ts (rewrites, for the browser) and by
// server components (direct calls, since the rewrite only matters for :3000 <-> browser).
export const SERVICE_URLS = {
  user: process.env.USER_SERVICE_URL ?? "http://localhost:8081",
  occupancy: process.env.OCCUPANCY_SERVICE_URL ?? "http://localhost:8083",
  reservation: process.env.RESERVATION_SERVICE_URL ?? "http://localhost:8084",
  review: process.env.REVIEW_SERVICE_URL ?? "http://localhost:8085",
} as const;
