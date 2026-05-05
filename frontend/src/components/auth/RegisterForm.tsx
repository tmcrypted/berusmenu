"use client";

import Link from "next/link";
import { useState, type FormEvent } from "react";
import { AuthShell } from "./AuthShell";

export function RegisterForm() {
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setError(null);
    const form = e.currentTarget;
    const password = (form.elements.namedItem("password") as HTMLInputElement)
      .value;
    const confirm = (
      form.elements.namedItem("confirmPassword") as HTMLInputElement
    ).value;
    if (password !== confirm) {
      setError("Пароли не совпадают.");
      return;
    }
    setPending(true);
    // TODO: интеграция с API регистрации
    setTimeout(() => setPending(false), 400);
  }

  return (
    <AuthShell
      title="Регистрация"
      subtitle="Создайте аккаунт, чтобы находить смены на ПВЗ или искать подмену."
      footer={
        <>
          Уже есть аккаунт?{" "}
          <Link
            href="/login"
            className="font-medium text-[var(--accent)] hover:text-[var(--accent-hover)]"
          >
            Войти
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="flex flex-col gap-4">
        {error ? (
          <p
            className="rounded-lg bg-red-500/10 px-3 py-2 text-sm text-red-600 dark:text-red-400"
            role="alert"
          >
            {error}
          </p>
        ) : null}
        <div>
          <label
            htmlFor="register-name"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Имя
          </label>
          <input
            id="register-name"
            name="name"
            type="text"
            autoComplete="name"
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2.5 text-[var(--foreground)] outline-none ring-[var(--accent)] transition placeholder:text-[var(--muted)] focus:ring-2"
            placeholder="Как к вам обращаться"
          />
        </div>
        <div>
          <label
            htmlFor="register-email"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Email
          </label>
          <input
            id="register-email"
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
            htmlFor="register-password"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Пароль
          </label>
          <input
            id="register-password"
            name="password"
            type="password"
            autoComplete="new-password"
            required
            minLength={8}
            className="w-full rounded-lg border border-[var(--border)] bg-[var(--background)] px-3 py-2.5 text-[var(--foreground)] outline-none ring-[var(--accent)] transition placeholder:text-[var(--muted)] focus:ring-2"
            placeholder="Не менее 8 символов"
          />
        </div>
        <div>
          <label
            htmlFor="register-confirm"
            className="mb-1.5 block text-sm font-medium text-[var(--foreground)]"
          >
            Повторите пароль
          </label>
          <input
            id="register-confirm"
            name="confirmPassword"
            type="password"
            autoComplete="new-password"
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
          {pending ? "Регистрация…" : "Зарегистрироваться"}
        </button>
      </form>
    </AuthShell>
  );
}
