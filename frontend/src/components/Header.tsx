"use client";

import Link from "next/link";
import { useResolvedAuthSnapshot } from "@/components/auth/AuthStoreHydrator";
import { profileCompletionPath } from "@/lib/profile";
import { AccountMenu } from "./AccountMenu";

const headerChrome =
  "inline-flex min-h-10 shrink-0 items-center justify-center rounded-xl sm:min-h-11";
const authSlot = "w-[12.5rem] sm:w-[15.5rem]";

export function Header() {
  const { status, user } = useResolvedAuthSnapshot();

  const isBootstrapping = status === "loading";
  const isAuthenticated = status !== "anonymous" && !isBootstrapping;
  const homeHref =
    status === "authenticated_profile_incomplete" && user
      ? profileCompletionPath(user.role)
      : status === "authenticated_unverified"
        ? "/verify-email"
        : "/";
  const ready = status === "authenticated_ready" && user;
  const navLinks = ready
    ? user.role === "owner"
      ? [
          { href: "/owner/pvz", label: "ПВЗ" },
          { href: "/owner/vacancies", label: "Вакансии" },
          { href: "/assignments", label: "Назначения" },
        ]
      : user.role === "moderator"
        ? [
            { href: "/moderation/pvz", label: "ПВЗ" },
            { href: "/moderation/matching", label: "Смены" },
          ]
        : [
            { href: "/vacancies", label: "Лента" },
            { href: "/responses", label: "Отклики" },
            { href: "/assignments", label: "Назначения" },
          ]
    : [];

  return (
    <header className="fixed inset-x-0 top-0 z-50 flex h-[var(--header-height)] items-center border-b border-[var(--border)] bg-[color-mix(in_srgb,var(--background)_88%,transparent)] backdrop-blur-xl">
      <div className="mx-auto flex h-full w-full max-w-7xl items-center justify-between gap-3 px-4 sm:gap-4 sm:px-6 lg:px-8">
        <Link href={homeHref} className="group min-w-0 shrink">
          <span className="inline-flex items-baseline text-sm font-semibold tracking-tight text-[var(--foreground)] transition group-hover:text-[var(--accent)] sm:text-base">
            <span className="truncate">Беру Смену</span>
            <span className="font-medium text-[var(--muted)]">.ру</span>
          </span>
        </Link>
        {navLinks.length > 0 ? (
          <nav className="hidden min-w-0 flex-1 items-center justify-center gap-1 md:flex" aria-label="Основные разделы">
            {navLinks.map((link) => (
              <Link key={link.href} href={link.href} className="rounded-xl px-3 py-2 text-sm font-semibold text-[var(--muted)] transition hover:bg-[var(--muted-bg)] hover:text-[var(--foreground)]">
                {link.label}
              </Link>
            ))}
          </nav>
        ) : null}
        <nav className="flex shrink-0 items-center" aria-label="Управление аккаунтом">
          <div className={authSlot}>
            {isAuthenticated && user ? (
              <AccountMenu user={user} status={status} />
            ) : isBootstrapping ? (
              <div
                className="h-10 w-full rounded-xl border border-[var(--border)] bg-[var(--surface)]/80 sm:h-11"
                aria-hidden="true"
              />
            ) : (
              <div className="flex w-full items-center justify-end gap-2 sm:gap-3">
                <Link
                  href="/login"
                  className={`${headerChrome} whitespace-nowrap px-2 text-xs font-medium text-[var(--foreground)] transition hover:bg-[var(--muted-bg)] sm:px-3.5 sm:text-sm`}
                >
                  Войти
                </Link>
                <Link
                  href="/register"
                  className={`${headerChrome} whitespace-nowrap bg-[var(--accent)] px-3 text-xs font-semibold text-white shadow-[var(--shadow-soft)] transition hover:bg-[var(--accent-hover)] sm:px-4 sm:text-sm`}
                >
                  Регистрация
                </Link>
              </div>
            )}
          </div>
        </nav>
      </div>
    </header>
  );
}
