import type { ReactNode } from "react";

type AuthShellProps = {
  title: string;
  subtitle?: string;
  children: ReactNode;
  footer?: ReactNode;
};

export function AuthShell({ title, subtitle, children, footer }: AuthShellProps) {
  return (
    <div className="relative isolate flex min-h-[calc(100vh_-_var(--header-height))] w-full overflow-hidden px-4 py-8 sm:px-6 sm:py-12 lg:px-8">
      <div className="absolute inset-0 -z-10 bg-[var(--background)]" />
      <div className="mx-auto flex w-full max-w-xl flex-col justify-center">
        <div className="w-full rounded-[2rem] border border-[var(--border)] bg-[color-mix(in_srgb,var(--card)_94%,transparent)] p-6 shadow-[var(--shadow-soft)] backdrop-blur-sm sm:p-8">
          <h2 className="text-2xl font-semibold tracking-tight text-[var(--foreground)] sm:text-3xl">
            {title}
          </h2>
          {subtitle ? (
            <p className="mt-3 text-sm leading-7 text-[var(--muted)] sm:text-base">
              {subtitle}
            </p>
          ) : null}
          <div className="mt-8">{children}</div>
        </div>
        {footer ? (
          <div className="mt-6 text-center text-sm text-[var(--muted)]">{footer}</div>
        ) : null}
      </div>
    </div>
  );
}
