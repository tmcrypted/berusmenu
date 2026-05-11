import type { PublicUser, UserRole } from "@/types/auth";

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

export function normalizePublicUser(raw: unknown): PublicUser {
  const r = (raw ?? {}) as Record<string, unknown>;
  const id = Number(r.id);
  const avatar =
    r.avatar_url === null || r.avatar_url === undefined
      ? null
      : typeof r.avatar_url === "string"
        ? r.avatar_url
        : String(r.avatar_url);
  const telegram =
    r.telegram_username === null || r.telegram_username === undefined
      ? null
      : String(r.telegram_username);
  const max =
    r.max_username === null || r.max_username === undefined
      ? null
      : String(r.max_username);

  return {
    id: Number.isFinite(id) ? id : 0,
    role: typeof r.role === "string" ? API_ROLE_MAP[r.role] ?? "unspecified" : "unspecified",
    last_name: typeof r.last_name === "string" ? r.last_name : "",
    first_name: typeof r.first_name === "string" ? r.first_name : "",
    middle_name: r.middle_name === null || r.middle_name === undefined ? null : String(r.middle_name),
    birth_date: r.birth_date === null || r.birth_date === undefined ? null : String(r.birth_date),
    avatar_url: avatar,
    telegram_username: telegram,
    max_username: max,
    created_at: typeof r.created_at === "string" ? r.created_at : "",
    updated_at: typeof r.updated_at === "string" ? r.updated_at : "",
  };
}
