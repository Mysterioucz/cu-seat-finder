import type { ApiErrorBody } from "./types";

export class ApiError extends Error {
  code: string;
  status: number;
  constructor(status: number, body: ApiErrorBody) {
    super(body.message || body.error || "Request failed");
    this.code = body.error ?? "unknown";
    this.status = status;
  }
}
