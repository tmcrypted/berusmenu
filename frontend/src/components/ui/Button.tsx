"use client";

import type { ButtonHTMLAttributes, ReactNode } from "react";
import { cn } from "@/lib/cn";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & {
  children: ReactNode;
  variant?: "primary" | "secondary" | "ghost";
  fullWidth?: boolean;
};

const variantClasses = {
  primary:
    "bg-[var(--accent)] text-white shadow-[var(--shadow-soft)] hover:bg-[var(--accent-hover)]",
  secondary:
    "border border-[var(--border)] bg-[var(--card)] text-[var(--foreground)] hover:bg-[var(--muted-bg)]",
  ghost: "text-[var(--foreground)] hover:bg-[var(--muted-bg)]",
};

export function Button({
  className,
  children,
  variant = "primary",
  fullWidth = false,
  ...props
}: Props) {
  return (
    <button
      className={cn(
        "inline-flex min-h-11 items-center justify-center rounded-2xl px-4 py-2.5 text-sm font-semibold transition disabled:cursor-not-allowed disabled:opacity-60",
        variantClasses[variant],
        fullWidth && "w-full",
        className,
      )}
      {...props}
    >
      {children}
    </button>
  );
}
