import type { NextConfig } from "next";
import { SERVICE_URLS } from "./src/lib/serviceUrls";

// Per .planning/API-CONTRACT.md topology table. Env-overridable for non-dev deploys.
const USER_SERVICE_URL = SERVICE_URLS.user;
const OCCUPANCY_SERVICE_URL = SERVICE_URLS.occupancy;
const RESERVATION_SERVICE_URL = SERVICE_URLS.reservation;
const REVIEW_SERVICE_URL = SERVICE_URLS.review;

// /api/libraries and /api/rooms are route handlers (gRPC facade, src/lib/libraryGrpc.ts), not rewrites.
const nextConfig: NextConfig = {
  serverExternalPackages: ["@grpc/grpc-js", "@grpc/proto-loader"],
  async rewrites() {
    return [
      { source: "/api/auth", destination: `${USER_SERVICE_URL}/api/auth` },
      { source: "/api/auth/:path*", destination: `${USER_SERVICE_URL}/api/auth/:path*` },
      { source: "/api/users", destination: `${USER_SERVICE_URL}/api/users` },
      { source: "/api/users/:path*", destination: `${USER_SERVICE_URL}/api/users/:path*` },

      { source: "/api/occupancy", destination: `${OCCUPANCY_SERVICE_URL}/api/occupancy` },
      { source: "/api/occupancy/:path*", destination: `${OCCUPANCY_SERVICE_URL}/api/occupancy/:path*` },

      { source: "/api/reservations", destination: `${RESERVATION_SERVICE_URL}/api/reservations` },
      { source: "/api/reservations/:path*", destination: `${RESERVATION_SERVICE_URL}/api/reservations/:path*` },

      { source: "/api/reviews", destination: `${REVIEW_SERVICE_URL}/api/reviews` },
      { source: "/api/reviews/:path*", destination: `${REVIEW_SERVICE_URL}/api/reviews/:path*` },
    ];
  },
};

export default nextConfig;
