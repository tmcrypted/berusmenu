import type { AuthStatus, User, UserRole } from "@/types/auth";

export function isProfileComplete(user: User) {
  const name = user.full_name?.trim() ?? "";
  const phone = user.phone?.trim() ?? "";
  return Boolean(name && phone);
}

export function profileCompletionPath(role: UserRole) {
  if (role === "owner") {
    return "/complete-profile/owner";
  }

  return "/complete-profile/worker";
}

/** Ссылка «Профиль» / онбординг в зависимости от состояния auth. */
export function getProfileNavHref(status: AuthStatus, user: User | null): string {
  if (!user) {
    return "/login";
  }
  if (status === "authenticated_unverified") {
    return "/verify-email";
  }
  if (status === "authenticated_profile_incomplete") {
    return profileCompletionPath(user.role);
  }
  return "/profile";
}

export function roleLabel(role: UserRole) {
  switch (role) {
    case "owner":
      return "Владелец ПВЗ";
    case "moderator":
      return "Модератор";
    default:
      return "Исполнитель";
  }
}
