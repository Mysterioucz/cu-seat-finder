// The API Gateway is the only backend address the frontend knows (ADR-008): it routes
// every /api/* path to the owning service and translates library calls to gRPC.
// Used by next.config.ts (rewrites, for the browser) and by serverApi (server components).
export const GATEWAY_URL = process.env.GATEWAY_URL ?? "http://localhost:8080";
