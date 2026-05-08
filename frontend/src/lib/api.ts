import { readCookie } from "./cookies";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

type ApiErrorShape = {
  error?: {
    code?: string;
    message?: string;
    request_id?: string;
  };
};

type RequestOptions = {
  method?: "GET" | "POST" | "PATCH" | "PUT" | "DELETE";
  body?: unknown;
};

export class ApiError extends Error {
  status: number;
  code: string;
  requestId?: string;

  constructor(status: number, code: string, message: string, requestId?: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.requestId = requestId;
  }
}

export async function apiRequest<T>(path: string, options: RequestOptions = {}) {
  const method = options.method ?? "GET";
  const headers = new Headers();
  const isUnsafe = method !== "GET";

  if (options.body !== undefined) {
    headers.set("Content-Type", "application/json");
  }

  if (isUnsafe) {
    const csrfToken = readCookie("berusmenu_csrf");
    if (csrfToken) {
      headers.set("X-CSRF-Token", csrfToken);
    }
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method,
    credentials: "include",
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });

  if (response.status === 204) {
    return undefined as T;
  }

  const text = await response.text();
  const data = text ? (JSON.parse(text) as T | ApiErrorShape) : ({} as T);

  if (!response.ok) {
    const errorData = data as ApiErrorShape;
    throw new ApiError(
      response.status,
      errorData.error?.code ?? "unknown_error",
      errorData.error?.message ?? "Request failed",
      errorData.error?.request_id,
    );
  }

  return data as T;
}
