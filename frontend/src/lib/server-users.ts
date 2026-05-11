import "server-only";

import { serverApiRequest, type ServerApiResult } from "@/lib/server-api";
import { normalizePublicUser } from "@/lib/users";
import type { PublicUser } from "@/types/auth";

export async function getPublicUser(id: number): Promise<ServerApiResult<PublicUser>> {
  const result = await serverApiRequest<unknown>(`/v1/users/${id}`);
  return result.ok ? { ok: true, data: normalizePublicUser(result.data) } : result;
}
