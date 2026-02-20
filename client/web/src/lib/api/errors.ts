import type { ApiError as ApiErrorBody } from '$lib/api/schema/app.js';

// ─── Base: API Error ─────────────────────────────────────────────────────────
// Represents any non-2xx response from the backend.
// Carries the HTTP status and the parsed error body when available.
export class ApiError extends Error {
	constructor(
		public readonly status: number,
		message: string,
		public readonly body?: ApiErrorBody
	) {
		super(message);
		this.name = 'ApiError';
	}
}

// ─── 401 Unauthorized ────────────────────────────────────────────────────────
export class AuthError extends ApiError {
	constructor(body?: ApiErrorBody) {
		super(401, body?.message ?? 'Unauthorized', body);
		this.name = 'AuthError';
	}
}

// ─── 403 Forbidden ───────────────────────────────────────────────────────────
export class ForbiddenError extends ApiError {
	constructor(body?: ApiErrorBody) {
		super(403, body?.message ?? 'Forbidden', body);
		this.name = 'ForbiddenError';
	}
}

// ─── 404 Not Found ───────────────────────────────────────────────────────────
export class NotFoundError extends ApiError {
	constructor(body?: ApiErrorBody) {
		super(404, body?.message ?? 'Not Found', body);
		this.name = 'NotFoundError';
	}
}

// ─── 409 Conflict ────────────────────────────────────────────────────────────
export class ConflictError extends ApiError {
	constructor(body?: ApiErrorBody) {
		super(409, body?.message ?? 'Conflict', body);
		this.name = 'ConflictError';
	}
}

// ─── 422 Unprocessable Entity ────────────────────────────────────────────────
export class UnprocessableError extends ApiError {
	constructor(body?: ApiErrorBody) {
		super(422, body?.message ?? 'Unprocessable Entity', body);
		this.name = 'UnprocessableError';
	}
}

// ─── 5xx Server Error ────────────────────────────────────────────────────────
export class ServerError extends ApiError {
	constructor(status: number, body?: ApiErrorBody) {
		super(status, body?.message ?? 'Internal Server Error', body);
		this.name = 'ServerError';
	}
}

// ─── Network / Transport Failures ────────────────────────────────────────────
// Covers fetch timeouts (AbortError), DNS failures, connection refused, etc.
export class NetworkError extends Error {
	constructor(message: string, public readonly cause?: unknown) {
		super(message);
		this.name = 'NetworkError';
	}
}

// ─── Response Validation Failures ────────────────────────────────────────────
// Thrown when the response body doesn't match the expected Zod schema.
// This is a programmer/contract error, not a user error.
export class ValidationError extends Error {
	constructor(message: string, public readonly cause?: unknown) {
		super(message);
		this.name = 'ValidationError';
	}
}

// ─── Factory: map HTTP status → typed error ──────────────────────────────────
export function createApiError(status: number, body?: ApiErrorBody): ApiError {
	switch (status) {
		case 401:
			return new AuthError(body);
		case 403:
			return new ForbiddenError(body);
		case 404:
			return new NotFoundError(body);
		case 409:
			return new ConflictError(body);
		case 422:
			return new UnprocessableError(body);
		default:
			if (status >= 500) return new ServerError(status, body);
			return new ApiError(status, body?.message ?? `API error: ${status}`, body);
	}
}