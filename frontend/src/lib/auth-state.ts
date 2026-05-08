import { normalizeApiUser } from "@/lib/api-user";
import { isProfileComplete } from "@/lib/profile";
import type { ApiUserWire, AuthStatus, User } from "@/types/auth";

export type AuthSnapshot = {
  user: User | null;
  status: AuthStatus;
  initialized: boolean;
  accessExpiresAt: string | null;
};

export function createLoadingAuthSnapshot(): AuthSnapshot {
  return {
    user: null,
    status: "loading",
    initialized: false,
    accessExpiresAt: null,
  };
}

export function createAnonymousAuthSnapshot(): AuthSnapshot {
  return {
    user: null,
    status: "anonymous",
    initialized: true,
    accessExpiresAt: null,
  };
}

export function resolveUserAuthSnapshot(
  user: User,
  accessExpiresAt?: string | null,
): AuthSnapshot {
  if (!user.email_verified_at) {
    return {
      user,
      status: "authenticated_unverified",
      initialized: true,
      accessExpiresAt: accessExpiresAt ?? null,
    };
  }

  if (isProfileComplete(user)) {
    return {
      user,
      status: "authenticated_ready",
      initialized: true,
      accessExpiresAt: accessExpiresAt ?? null,
    };
  }

  return {
    user,
    status: "authenticated_profile_incomplete",
    initialized: true,
    accessExpiresAt: accessExpiresAt ?? null,
  };
}

export function resolveApiAuthSnapshot(
  rawUser: ApiUserWire,
  accessExpiresAt?: string | null,
): AuthSnapshot {
  return resolveUserAuthSnapshot(normalizeApiUser(rawUser), accessExpiresAt);
}
