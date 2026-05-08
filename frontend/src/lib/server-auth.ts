import "server-only";

import { cookies } from "next/headers";
import {
  createAnonymousAuthSnapshot,
  createLoadingAuthSnapshot,
  type AuthSnapshot,
} from "@/lib/auth-state";
import { normalizeApiUser } from "@/lib/api-user";
import { resolveUserAuthSnapshot } from "@/lib/auth-state";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_BASE_URL ?? "http://localhost:8080";

function toCookieHeader(entries: Awaited<ReturnType<typeof cookies>>): string {
  return entries
    .getAll()
    .map(({ name, value }) => `${encodeURIComponent(name)}=${encodeURIComponent(value)}`)
    .join("; ");
}

export async function getServerAuthSnapshot(): Promise<AuthSnapshot> {
  const cookieStore = await cookies();
  const cookieHeader = toCookieHeader(cookieStore);

  if (!cookieHeader) {
    return createAnonymousAuthSnapshot();
  }

  const headers = new Headers({
    Accept: "application/json",
    Cookie: cookieHeader,
  });

  try {
    const response = await fetch(`${API_BASE_URL}/v1/auth/session`, {
      method: "GET",
      headers,
      cache: "no-store",
    });

    if (response.status === 401 || response.status === 403) {
      return createAnonymousAuthSnapshot();
    }

    if (!response.ok) {
      return createLoadingAuthSnapshot();
    }

    const user = normalizeApiUser(await response.json());
    return resolveUserAuthSnapshot(user);
  } catch {
    return createLoadingAuthSnapshot();
  }
}
