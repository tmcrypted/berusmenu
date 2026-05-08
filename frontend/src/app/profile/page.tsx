"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useResolvedAuthSnapshot } from "@/components/auth/AuthStoreHydrator";
import { Spinner } from "@/components/ui/Spinner";
import { profileCompletionPath, roleLabel } from "@/lib/profile";

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

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-2xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <h1 className="text-2xl font-semibold tracking-tight text-[var(--foreground)]">
            Профиль
          </h1>
          <dl className="mt-8 space-y-4 text-sm">
            <div>
              <dt className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--muted)]">
                Роль
              </dt>
              <dd className="mt-1 text-[var(--foreground)]">{roleLabel(user.role)}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--muted)]">
                Email
              </dt>
              <dd className="mt-1 text-[var(--foreground)]">{user.email}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--muted)]">
                Имя
              </dt>
              <dd className="mt-1 text-[var(--foreground)]">{user.full_name?.trim() || "—"}</dd>
            </div>
            <div>
              <dt className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--muted)]">
                Телефон
              </dt>
              <dd className="mt-1 text-[var(--foreground)]">{user.phone?.trim() || "—"}</dd>
            </div>
          </dl>
          <p className="mt-8 text-xs text-[var(--muted)]">
            Смена пароля и расширенные настройки — в следующих итерациях. Данные синхронизируются с auth-service.
          </p>
        </div>
      </section>
    </div>
  );
}
