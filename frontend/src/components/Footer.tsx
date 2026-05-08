"use client";

import { ThemeToggle } from "./ThemeToggle";

export function Footer() {
  return (
    <footer className="border-t border-[var(--border)] bg-[color-mix(in_srgb,var(--background)_92%,transparent)]">
      <div className="mx-auto flex w-full max-w-7xl items-center justify-between gap-3 px-4 py-4 sm:px-6 lg:px-8">
        <p className="text-xs text-[var(--muted)] sm:text-sm">Беру Смену.ру</p>
        <div className="flex items-center gap-3">
          <span className="text-xs text-[var(--muted)] sm:text-sm">Тема</span>
          <ThemeToggle />
        </div>
      </div>
    </footer>
  );
}
