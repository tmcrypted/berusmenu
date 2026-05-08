import { apiRequest } from "@/lib/api";
import type {
  BlockUserAndDeletePvzInput,
  BlockUserAndDeletePvzResponse,
  ModerationDecisionInput,
  Pvz,
  PvzFormInput,
  PvzListResponse,
  PvzStatus,
} from "@/types/inventory";

type PvzWire = Record<string, unknown>;

const STATUS_MAP: Record<string, PvzStatus> = {
  PVZ_STATUS_PENDING: "pending",
  PVZ_STATUS_APPROVED: "approved",
  PVZ_STATUS_REJECTED: "rejected",
  PVZ_STATUS_BLOCKED: "blocked",
  pending: "pending",
  approved: "approved",
  rejected: "rejected",
  blocked: "blocked",
};

function coerceStatus(value: unknown): PvzStatus {
  if (typeof value === "number" && Number.isFinite(value)) {
    switch (value) {
      case 1:
        return "pending";
      case 2:
        return "approved";
      case 3:
        return "rejected";
      case 4:
        return "blocked";
      default:
        return "unspecified";
    }
  }
  if (typeof value === "string" && /^\d+$/.test(value)) {
    return coerceStatus(Number(value));
  }
  if (typeof value === "string") {
    return STATUS_MAP[value] ?? "unspecified";
  }
  return "unspecified";
}

function numberOrNull(value: unknown): number | null {
  if (value === null || value === undefined || value === "") {
    return null;
  }
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : null;
}

function stringOrNull(value: unknown): string | null {
  if (value === null || value === undefined) {
    return null;
  }
  return typeof value === "string" ? value : String(value);
}

export function normalizePvz(raw: unknown): Pvz {
  if (!raw || typeof raw !== "object") {
    throw new Error("Invalid pvz payload");
  }
  const r = raw as PvzWire;
  const id = Number(r.id);
  const ownerId = Number(r.owner_id);
  if (!Number.isFinite(id) || !Number.isFinite(ownerId)) {
    throw new Error("Invalid pvz id payload");
  }

  return {
    id,
    owner_id: ownerId,
    status: coerceStatus(r.status),
    city: typeof r.city === "string" ? r.city : "",
    address: typeof r.address === "string" ? r.address : "",
    company_name: typeof r.company_name === "string" ? r.company_name : "",
    contact_phone: typeof r.contact_phone === "string" ? r.contact_phone : "",
    contact_telegram: stringOrNull(r.contact_telegram),
    description: stringOrNull(r.description),
    lat: numberOrNull(r.lat),
    lng: numberOrNull(r.lng),
    rejection_reason: stringOrNull(r.rejection_reason),
    moderated_by: numberOrNull(r.moderated_by),
    moderated_at: stringOrNull(r.moderated_at),
    created_at: typeof r.created_at === "string" ? r.created_at : "",
    updated_at: typeof r.updated_at === "string" ? r.updated_at : "",
  };
}

function cleanPayload(input: PvzFormInput): PvzFormInput {
  const payload: PvzFormInput = {
    city: input.city.trim(),
    address: input.address.trim(),
    company_name: input.company_name.trim(),
    contact_phone: input.contact_phone.trim(),
  };
  if (input.contact_telegram?.trim()) {
    payload.contact_telegram = input.contact_telegram.trim();
  }
  if (input.description?.trim()) {
    payload.description = input.description.trim();
  }
  if (typeof input.lat === "number" && Number.isFinite(input.lat)) {
    payload.lat = input.lat;
  }
  if (typeof input.lng === "number" && Number.isFinite(input.lng)) {
    payload.lng = input.lng;
  }
  return payload;
}

function normalizeList(raw: unknown): PvzListResponse {
  if (!raw || typeof raw !== "object") {
    return { items: [] };
  }
  const r = raw as Record<string, unknown>;
  const items = Array.isArray(r.items) ? r.items.map(normalizePvz) : [];
  return {
    items,
    next_page_token:
      typeof r.next_page_token === "string" && r.next_page_token ? r.next_page_token : undefined,
  };
}

export async function createPvz(input: PvzFormInput) {
  return normalizePvz(
    await apiRequest<unknown>("/v1/pvz", {
      method: "POST",
      body: cleanPayload(input),
    }),
  );
}

export async function updatePvz(id: number, input: PvzFormInput) {
  return normalizePvz(
    await apiRequest<unknown>(`/v1/pvz/${id}`, {
      method: "PATCH",
      body: cleanPayload(input),
    }),
  );
}

export async function listOwnerPvz(pageToken?: string) {
  const query = pageToken ? `?page_size=20&page_token=${encodeURIComponent(pageToken)}` : "?page_size=20";
  return normalizeList(await apiRequest<unknown>(`/v1/pvz${query}`));
}

export async function getPvz(id: number) {
  return normalizePvz(await apiRequest<unknown>(`/v1/pvz/${id}`));
}

export async function listPendingPvz(pageToken?: string) {
  const query = pageToken ? `?page_size=20&page_token=${encodeURIComponent(pageToken)}` : "?page_size=20";
  return normalizeList(await apiRequest<unknown>(`/v1/moderation/pvz${query}`));
}

export async function approvePvz(id: number, input: ModerationDecisionInput) {
  return normalizePvz(
    await apiRequest<unknown>(`/v1/moderation/pvz/${id}/approve`, {
      method: "POST",
      body: { comment: input.comment?.trim() || undefined },
    }),
  );
}

export async function rejectPvz(id: number, input: ModerationDecisionInput) {
  return normalizePvz(
    await apiRequest<unknown>(`/v1/moderation/pvz/${id}/reject`, {
      method: "POST",
      body: {
        reason: input.reason?.trim(),
        comment: input.comment?.trim() || undefined,
      },
    }),
  );
}

export async function blockPvz(id: number, input: ModerationDecisionInput) {
  return normalizePvz(
    await apiRequest<unknown>(`/v1/moderation/pvz/${id}/block`, {
      method: "POST",
      body: { comment: input.comment?.trim() || undefined },
    }),
  );
}

export async function blockUserAndDeletePvzRequests(ownerId: number, input: BlockUserAndDeletePvzInput) {
  return apiRequest<BlockUserAndDeletePvzResponse>(
    `/v1/moderation/users/${ownerId}/block-and-delete-pvz`,
    {
      method: "POST",
      body: {
        reason: input.reason.trim(),
        block_duration_minutes: input.block_duration_minutes,
        comment: input.comment?.trim() || undefined,
      },
    },
  );
}
