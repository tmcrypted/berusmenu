"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { logout } from "@/lib/auth";
import { getProfileNavHref, roleLabel } from "@/lib/profile";
import type { AuthStatus, User } from "@/types/auth";
import { useAuthStore } from "@/store/auth-store";
import { Button } from "./ui/Button";

type Props = {
  user: User;
  status: AuthStatus;
};

export function AccountMenu({ user, status }: Props) {
  const router = useRouter();
  const clearAuth = useAuthStore((s) => s.clearAuth);
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) {
      return;
    }
    function onPointerDown(e: MouseEvent) {
      if (rootRef.current && !rootRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", onPointerDown);
    document.addEventListener("keydown", onKey);
    return () => {
      document.removeEventListener("mousedown", onPointerDown);
      document.removeEventListener("keydown", onKey);
    };
  }, [open]);

  async function handleLogout() {
    setOpen(false);
    try {
      await logout();
    } finally {
      clearAuth();
      router.push("/login");
    }
  }

  const profileHref = getProfileNavHref(status, user);
  const isOwner = user.role === "owner";
  const isModerator = user.role === "moderator";

  return (
    <div className="relative w-full" ref={rootRef}>
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className="inline-flex w-full min-h-10 items-center gap-1.5 rounded-xl border border-[var(--border)] bg-[var(--surface)] px-2.5 py-2 text-left text-xs text-[var(--foreground)] transition hover:bg-[var(--muted-bg)] sm:min-h-11 sm:px-3.5 sm:text-sm"
        aria-expanded={open}
        aria-haspopup="true"
      >
        <span className="min-w-0 flex-1 truncate">
          <span className="block truncate font-medium">{roleLabel(user.role)}</span>
          <span className="block truncate text-[var(--muted)]">{user.email}</span>
        </span>
        <span className="shrink-0 text-[var(--muted)]" aria-hidden>
          {open ? "▲" : "▼"}
        </span>
      </button>

      {open ? (
        <div
          className="absolute right-0 z-[100] mt-2 w-[min(22rem,calc(100vw-1.5rem))] origin-top-right rounded-2xl border border-[var(--border)] bg-[var(--card)] p-4 shadow-[var(--shadow-soft)]"
          role="menu"
        >
          <div className="border-b border-[var(--border)] pb-3">
            <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">
              Профиль
            </p>
            <p className="mt-2 text-sm font-medium text-[var(--foreground)]">
              {user.full_name?.trim() || "—"}
            </p>
            <p className="mt-1 truncate text-sm text-[var(--muted)]">{user.email}</p>
            <p className="mt-1 text-sm text-[var(--muted)]">{user.phone?.trim() || "—"}</p>
            <Link
              href={profileHref}
              onClick={() => setOpen(false)}
              className="mt-3 inline-flex text-sm font-semibold text-[var(--accent)] hover:text-[var(--accent-hover)]"
              role="menuitem"
            >
              {status === "authenticated_profile_incomplete"
                ? "Завершить профиль"
                : status === "authenticated_unverified"
                  ? "Подтвердить почту"
                  : "Страница профиля"}
            </Link>
          </div>

          {isOwner || isModerator ? (
            <div className="border-b border-[var(--border)] py-3">
              <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">
                {isOwner ? "ПВЗ" : "Модерация"}
              </p>
              <Link
                href={isOwner ? "/owner/pvz" : "/moderation/pvz"}
                onClick={() => setOpen(false)}
                className="mt-3 block"
              >
                <Button type="button" variant="secondary" fullWidth className="rounded-xl text-sm">
                  {isOwner ? "Мои ПВЗ" : "Очередь ПВЗ"}
                </Button>
              </Link>
              {isOwner ? (
                <Link href="/owner/pvz/new" onClick={() => setOpen(false)} className="mt-3 block">
                  <Button type="button" fullWidth className="rounded-xl text-sm">
                    Новый ПВЗ
                  </Button>
                </Link>
              ) : null}
            </div>
          ) : null}

          <div className="pt-3">
            <button
              type="button"
              role="menuitem"
              className="w-full rounded-xl px-3 py-2.5 text-left text-sm font-semibold text-[var(--danger)] transition hover:bg-[var(--danger-bg)]"
              onClick={() => void handleLogout()}
            >
              Выйти
            </button>
          </div>
        </div>
      ) : null}
    </div>
  );
}
