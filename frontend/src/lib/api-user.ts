import type { User, UserRole } from "@/types/auth";

const API_ROLE_MAP: Record<string, UserRole> = {
  USER_ROLE_UNSPECIFIED: "unspecified",
  USER_ROLE_OWNER: "owner",
  USER_ROLE_WORKER: "worker",
  USER_ROLE_MODERATOR: "moderator",
  unspecified: "unspecified",
  owner: "owner",
  worker: "worker",
  moderator: "moderator",
};

function coerceRole(value: unknown): UserRole {
  if (typeof value === "number" && Number.isFinite(value)) {
    switch (value) {
      case 1:
        return "owner";
      case 2:
        return "worker";
      case 3:
        return "moderator";
      default:
        return "unspecified";
    }
  }
  if (typeof value === "string" && /^\d+$/.test(value)) {
    return coerceRole(Number(value));
  }
  if (typeof value === "string") {
    return API_ROLE_MAP[value] ?? "unspecified";
  }
  return "unspecified";
}

/** Приводит ответ gateway (protojson) к типу `User`, ожидаемому в UI. */
export function normalizeApiUser(raw: unknown): User {
  if (!raw || typeof raw !== "object") {
    throw new Error("Invalid user payload");
  }
  const r = raw as Record<string, unknown>;

  const idRaw = r.id;
  const id = typeof idRaw === "number" ? idRaw : Number(idRaw);
  if (!Number.isFinite(id)) {
    throw new Error("Invalid user id");
  }

  const email = typeof r.email === "string" ? r.email : "";
  const full_name = typeof r.full_name === "string" ? r.full_name : "";
  const phone =
    r.phone === null || r.phone === undefined
      ? null
      : typeof r.phone === "string"
        ? r.phone
        : String(r.phone);
  const avatar_url =
    r.avatar_url === null || r.avatar_url === undefined
      ? null
      : typeof r.avatar_url === "string"
        ? r.avatar_url
        : String(r.avatar_url);

  const ts = (v: unknown) => (typeof v === "string" ? v : "");

  const email_verified_at =
    r.email_verified_at === null ||
    r.email_verified_at === undefined ||
    r.email_verified_at === ""
      ? null
      : String(r.email_verified_at);

  return {
    id,
    email,
    email_verified_at,
    role: coerceRole(r.role),
    full_name,
    phone,
    avatar_url,
    created_at: ts(r.created_at),
    updated_at: ts(r.updated_at),
  };
}
