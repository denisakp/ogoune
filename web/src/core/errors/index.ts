export abstract class ApiError extends Error {
  abstract readonly kind: string;

  /** Backend-supplied machine code. Optional. */
  public code?: string;

  /** `Retry-After` value in seconds (populated on 429). Optional. */
  public retryAfterSec?: number;

  constructor(
    message: string,
    public readonly status: number,
    public readonly traceId?: string,
  ) {
    super(message);
    this.name = this.constructor.name;
  }
}

export class ValidationError extends ApiError {
  readonly kind = "validation";

  constructor(
    message: string,
    public readonly fieldErrors: Record<string, string[]>,
    public readonly traceId?: string,
  ) {
    super(message, 400, traceId);
  }
}

export class UnauthorizedError extends ApiError {
  readonly kind = "unauthorized";
  constructor(message = "Authentication required", traceId?: string) {
    super(message, 401, traceId);
  }
}

export class ForbiddenError extends ApiError {
  readonly kind = "forbidden";

  constructor(message = "Insufficient permissions", traceId?: string) {
    super(message, 403, traceId);
  }
}

export class NotFoundError extends ApiError {
  readonly kind = "not_found";
  constructor(message = "Resource not found", traceId?: string) {
    super(message, 404, traceId);
  }
}

export class ConflictError extends ApiError {
  readonly kind = "conflict";
  constructor(message: string, traceId?: string) {
    super(message, 409, traceId);
  }
}

export class ServerError extends ApiError {
  readonly kind = "server";
  constructor(message = "Server error", status = 500, traceId?: string) {
    super(message, status, traceId);
  }
}

export class NetworkError extends ApiError {
  readonly kind = "network";
  constructor(message = "Network request failed") {
    super(message, 503);
  }
}
