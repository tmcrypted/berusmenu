"use client";

import { useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useResolvedAuthSnapshot } from "@/components/auth/AuthStoreHydrator";
import { UserAvatar } from "@/components/profile/UserAvatar";
import { Button } from "@/components/ui/Button";
import { Spinner } from "@/components/ui/Spinner";
import { ageLabel, displayName, profileCompletionPath, roleLabel } from "@/lib/profile";

export default function ProfilePage() {
  const router = useRouter();
  const { status, user, initialized } = useResolvedAuthSnapshot();

  useEffect(() => {
    if (!initialized || status === "loading") {
      return;
    }

    if (status === "anonymous") {
      router.replace("/login");
      return;
    }

    if (status === "authenticated_unverified") {
      router.replace("/verify-email");
      return;
    }

    if (status === "authenticated_profile_incomplete" && user) {
      router.replace(profileCompletionPath(user.role));
    }
  }, [initialized, router, status, user]);

  if (!initialized || status === "loading") {
    return (
      <div className="flex min-h-[calc(100vh_-_var(--header-height))] items-center justify-center">
        <Spinner />
      </div>
    );
  }

  if (
    status !== "authenticated_ready" ||
    !user
  ) {
    return null;
  }

  const name = displayName(user, "Профиль");
  const actions =
    user.role === "owner"
      ? [
          { href: "/owner/pvz", label: "Мои ПВЗ" },
          { href: "/owner/vacancies", label: "Мои вакансии" },
          { href: "/assignments", label: "Мои назначения" },
        ]
      : user.role === "moderator"
        ? [
            { href: "/moderation/pvz", label: "Модерация ПВЗ" },
            { href: "/moderation/matching", label: "Модерация смен" },
          ]
        : [
            { href: "/vacancies", label: "Лента смен" },
            { href: "/responses", label: "Мои отклики" },
            { href: "/assignments", label: "Мои назначения" },
          ];

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center sm:justify-between">
            <div className="flex min-w-0 items-center gap-4">
              <UserAvatar name={name} src={user.avatar_url} size="lg" />
              <div className="min-w-0">
                <p className="text-sm font-semibold text-[var(--accent)]">{roleLabel(user.role)}</p>
                <h1 className="mt-1 truncate text-3xl font-semibold tracking-tight text-[var(--foreground)]">{name}</h1>
                <p className="mt-2 text-sm text-[var(--muted)]">{ageLabel(user.birth_date)}</p>
              </div>
            </div>
          </div>
          <div className="mt-8 grid gap-4 md:grid-cols-2">
            <Info label="Email" value={user.email} />
            <Info label="Телефон" value={user.phone?.trim() || "Не указан"} />
            <Info label="Фамилия" value={user.last_name || "Не указана"} />
            <Info label="Имя" value={user.first_name || "Не указано"} />
            <Info label="Отчество" value={user.middle_name?.trim() || "Не указано"} />
            <Info label="Дата рождения" value={user.birth_date || "Не указана"} />
            <Info label="Telegram" value={formatUsername(user.telegram_username)} />
            <Info label="Max" value={formatUsername(user.max_username)} />
          </div>
        </div>

        <div className="mt-6 grid gap-4 md:grid-cols-3">
          {actions.map((action) => (
            <Link key={action.href} href={action.href} className="rounded-[1.25rem] border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)] transition hover:border-[var(--border-strong)]">
              <p className="text-lg font-semibold text-[var(--foreground)]">{action.label}</p>
              <Button type="button" variant="secondary" className="mt-4 rounded-xl">
                Открыть
              </Button>
            </Link>
          ))}
        </div>
      </section>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-2xl bg-[var(--surface)] p-4">
      <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">{label}</p>
      <p className="mt-2 text-sm font-medium text-[var(--foreground)]">{value}</p>
    </div>
  );
}

function formatUsername(value?: string | null) {
  const trimmed = value?.trim();
  return trimmed ? `@${trimmed.replace(/^@+/, "")}` : "Не указан";
}
