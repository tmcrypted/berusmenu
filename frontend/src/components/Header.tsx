import Link from "next/link";

export function Header() {
  return (
    <header className="sticky top-0 z-50 border-b border-[var(--border)] bg-[var(--background)]/90 backdrop-blur-md">
      <div className="mx-auto flex max-w-6xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
        <Link href="/" className="group flex items-baseline gap-2">
          <span className="text-lg font-semibold tracking-tight text-[var(--accent)] transition group-hover:text-[var(--accent-hover)]">
            Беру Смену
          </span>
          <span className="hidden text-xs text-[var(--muted)] sm:inline">
            berusmenu.ru
          </span>
        </Link>
        <nav className="flex items-center gap-2 sm:gap-3">
          <Link
            href="/login"
            className="rounded-lg px-3 py-2 text-sm font-medium text-[var(--foreground)] transition hover:bg-[var(--muted-bg)]"
          >
            Войти
          </Link>
          <Link
            href="/register"
            className="rounded-lg bg-[var(--accent)] px-3 py-2 text-sm font-medium text-white transition hover:bg-[var(--accent-hover)]"
          >
            Регистрация
          </Link>
        </nav>
      </div>
    </header>
  );
}
