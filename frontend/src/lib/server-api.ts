import "server-only";

import { cookies } from "next/headers";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

export type ServerApiResult<T> =
  | { ok: true; data: T }
  | { ok: false; status: number; message: string };

async function cookieHeader() {
  const cookieStore = await cookies();
  return cookieStore
    .getAll()
    .map(({ name, value }) => `${encodeURIComponent(name)}=${encodeURIComponent(value)}`)
    .join("; ");
}

export async function serverApiRequest<T>(path: string): Promise<ServerApiResult<T>> {
  const headers = new Headers({ Accept: "application/json" });
  const cookie = await cookieHeader();
  if (cookie) {
    headers.set("Cookie", cookie);
  }

  try {
    const response = await fetch(`${API_BASE_URL}${path}`, {
      method: "GET",
      headers,
      cache: "no-store",
    });
    const text = await response.text();
    const data = text ? JSON.parse(text) : {};

    if (!response.ok) {
      const message =
        data?.error?.message && typeof data.error.message === "string"
          ? data.error.message
          : "Request failed";
      return { ok: false, status: response.status, message };
    }

    return { ok: true, data: data as T };
  } catch {
    return { ok: false, status: 503, message: "API is unavailable" };
  }
}
