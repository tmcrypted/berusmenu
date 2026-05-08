import type { ReactNode } from "react";
import { cn } from "@/lib/cn";

type Props = {
  children: ReactNode;
  tone?: "info" | "error" | "success";
};

const toneClasses = {
  info: "border-[var(--border-strong)] bg-[var(--info-soft)] text-[var(--foreground)]",
  error: "border-[var(--danger-soft)] bg-[var(--danger-bg)] text-[var(--danger)]",
  success:
    "border-[var(--success-soft)] bg-[var(--success-bg)] text-[var(--success)]",
};

export function Alert({ children, tone = "info" }: Props) {
  return (
    <div
      className={cn(
        "rounded-2xl border px-4 py-3 text-sm leading-6",
        toneClasses[tone],
      )}
    >
      {children}
    </div>
  );
}
