import type { NextConfig } from "next";
import { GATEWAY_URL } from "./src/lib/gatewayUrl";

// Every /api/* call goes to the API Gateway (ADR-008), which knows where each service lives.
const nextConfig: NextConfig = {
  output: "standalone", // self-contained server for the Docker image
  async rewrites() {
    return [{ source: "/api/:path*", destination: `${GATEWAY_URL}/api/:path*` }];
  },
};

export default nextConfig;
