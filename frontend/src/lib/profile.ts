import type { AuthStatus, User, UserRole } from "@/types/auth";

export function isProfileComplete(user: User) {
  if (user.role === "moderator") {
    return true;
  }
  const lastName = user.last_name?.trim() ?? "";
  const firstName = user.first_name?.trim() ?? "";
  const birthDate = user.birth_date?.trim() ?? "";
  const phone = user.phone?.trim() ?? "";
  return Boolean(lastName && firstName && birthDate && phone);
}

export function displayName(user: Pick<User, "last_name" | "first_name" | "middle_name"> | null | undefined, fallback = "Пользователь") {
  const parts = [user?.last_name, user?.first_name, user?.middle_name].map((part) => part?.trim()).filter(Boolean);
  return parts.join(" ") || fallback;
}

export function shortDisplayName(user: Pick<User, "last_name" | "first_name" | "middle_name"> | null | undefined, fallback = "Пользователь") {
  const firstName = user?.first_name?.trim();
  const lastName = user?.last_name?.trim();
  const middleName = user?.middle_name?.trim();
  if (!firstName) {
    return displayName(user, fallback);
  }
  const initials = [lastName, middleName]
    .filter((part): part is string => Boolean(part))
    .map((part) => `${part[0].toUpperCase()}.`)
    .join("");
  return initials ? `${firstName} ${initials}` : firstName;
}

export function ageLabel(birthDate?: string | null) {
  if (!birthDate) {
    return "Возраст не указан";
  }
  const date = new Date(`${birthDate}T00:00:00`);
  if (Number.isNaN(date.getTime())) {
    return "Возраст не указан";
  }
  const now = new Date();
  let age = now.getFullYear() - date.getFullYear();
  const hadBirthday = now.getMonth() > date.getMonth() || (now.getMonth() === date.getMonth() && now.getDate() >= date.getDate());
  if (!hadBirthday) {
    age -= 1;
  }
  return age > 0 ? `${age} лет` : "Возраст не указан";
}

export function profileCompletionPath(role: UserRole) {
  if (role === "moderator") {
    return "/moderation/pvz";
  }
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
