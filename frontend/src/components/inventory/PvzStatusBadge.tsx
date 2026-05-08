import { cn } from "@/lib/cn";
import type { PvzStatus } from "@/types/inventory";

const STATUS_LABEL: Record<PvzStatus, string> = {
  pending: "На модерации",
  approved: "Одобрен",
  rejected: "Отклонен",
  blocked: "Заблокирован",
  unspecified: "Неизвестно",
};

const STATUS_CLASS: Record<PvzStatus, string> = {
  pending: "border-amber-300/70 bg-amber-100 text-amber-900",
  approved: "border-emerald-300/70 bg-emerald-100 text-emerald-900",
  rejected: "border-rose-300/70 bg-rose-100 text-rose-900",
  blocked: "border-zinc-300 bg-zinc-100 text-zinc-700",
  unspecified: "border-[var(--border)] bg-[var(--surface)] text-[var(--muted)]",
};

export function PvzStatusBadge({ status }: { status: PvzStatus }) {
  return (
    <span
      className={cn(
        "inline-flex min-h-7 items-center rounded-full border px-3 text-xs font-semibold",
        STATUS_CLASS[status],
      )}
    >
      {STATUS_LABEL[status]}
    </span>
  );
}

export function pvzStatusLabel(status: PvzStatus) {
  return STATUS_LABEL[status];
}
