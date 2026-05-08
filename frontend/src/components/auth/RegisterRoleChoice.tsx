"use client";

import Link from "next/link";
import { AuthGuard } from "./AuthGuard";
import { AuthShell } from "./AuthShell";

export function RegisterRoleChoice() {
  return (
    <AuthGuard mode="anonymous">
      <AuthShell
        title="Регистрация"
        subtitle="Выберите, как вы планируете пользоваться сервисом — дальше введите почту и пароль."
        footer={
          <>
            Уже есть аккаунт?{" "}
            <Link
              href="/login"
              className="font-semibold text-[var(--accent)] hover:text-[var(--accent-hover)]"
            >
              Войти
            </Link>
          </>
        }
      >
        <div className="grid gap-3 sm:gap-4">
          <Link
            href="/register/worker"
            className="flex min-h-14 items-center justify-center rounded-2xl bg-[var(--accent)] px-4 py-3.5 text-center text-sm font-semibold text-white shadow-[var(--shadow-soft)] transition hover:bg-[var(--accent-hover)] sm:min-h-[3.25rem] sm:text-base"
          >
            Я ищу работу
          </Link>
          <Link
            href="/register/owner"
            className="flex min-h-14 items-center justify-center rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-4 py-3.5 text-center text-sm font-semibold text-[var(--foreground)] transition hover:bg-[var(--muted-bg)] sm:min-h-[3.25rem] sm:text-base"
          >
            Я владелец ПВЗ
          </Link>
        </div>
      </AuthShell>
    </AuthGuard>
  );
}
