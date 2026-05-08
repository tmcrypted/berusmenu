import type { ReactNode } from "react";

type Props = {
  label: string;
  htmlFor: string;
  hint?: string;
  error?: string;
  children: ReactNode;
};

export function Field({ label, htmlFor, hint, error, children }: Props) {
  return (
    <div className="space-y-2">
      <label
        htmlFor={htmlFor}
        className="block text-sm font-medium text-[var(--foreground)]"
      >
        {label}
      </label>
      {children}
      {error ? (
        <p className="text-sm text-[var(--danger)]" role="alert">
          {error}
        </p>
      ) : hint ? (
        <p className="text-sm text-[var(--muted)]">{hint}</p>
      ) : null}
    </div>
  );
}
