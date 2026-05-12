import Link from "next/link";
import { UserAvatar } from "@/components/profile/UserAvatar";
import { ageLabel, displayName, roleLabel } from "@/lib/profile";
import type { PublicUser } from "@/types/auth";

type Props = {
  user?: PublicUser | null;
  fallbackId: number;
  compact?: boolean;
};

export function UserSummary({ user, fallbackId, compact = false }: Props) {
  const name = displayName(user, `Пользователь #${fallbackId}`);

  return (
    <Link href={`/profile/${fallbackId}`} className="group flex min-w-0 items-center gap-3">
      <UserAvatar name={name} src={user?.avatar_url} size={compact ? "sm" : "md"} />
      <div className="min-w-0">
        <p className="truncate text-sm font-semibold text-[var(--foreground)] group-hover:text-[var(--accent)]">
          {name}
        </p>
        <p className="mt-0.5 text-xs text-[var(--muted)]">
          {user ? roleLabel(user.role) : "Профиль недоступен"} · {ageLabel(user?.birth_date)}
        </p>
      </div>
    </Link>
  );
}
