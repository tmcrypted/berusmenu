"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { AuthShell } from "./AuthShell";

export function LoginForm() {
  const [pending, setPending] = useState(false);

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setPending(true);
    // TODO: интеграция с API авторизации
    setTimeout(() => setPending(false), 400);
  }

  return (
    <AuthShell
      title="Вход"
      subtitle="Войдите, чтобы откликаться на смены и размещать объявления."
      footer={
        <>
          Нет аккаунта?{" "}
          <Link
            href="/register"
            className="font-medium text-[var(--accent)] hover:text-[var(--accent-hover)]"
          >
            Зарегистрироваться
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        <div>
          <label
            htmlFor="login-email"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Email
          </label>
          <input
            id="login-email"
            name="email"
            type="email"
            autoComplete="email"
            required
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2.5 text-[var(--foreground)] outline-none ring-[var(--accent)] transition placeholder:text-[var(--muted)] focus:ring-2"
            placeholder="you@example.com"
          />
        </div>
        <div>
          <label
            htmlFor="login-password"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Пароль
          </label>
          <input
            id="login-password"
            name="password"
            type="password"
            autoComplete="current-password"
            required
            minLength={8}
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2.5 text-[var(--foreground)] outline-none ring-[var(--accent)] transition placeholder:text-[var(--muted)] focus:ring-2"
            placeholder="••••••••"
          />
        </div>
        <button
          type="submit"
          disabled={pending}
          className="mt-2 rounded-lg bg-[var(--accent)] px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-[var(--accent-hover)] disabled:opacity-60"
        >
          {pending ? "Вход…" : "Войти"}
        </button>
      </form>
    </AuthShell>
  );
}
