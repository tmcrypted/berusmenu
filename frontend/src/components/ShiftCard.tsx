import type { ShiftCard as ShiftCardType } from "@/types/shift";
import { Button } from "./ui/Button";

type Props = {
  shift: ShiftCardType;
};

export function ShiftCard({ shift }: Props) {
  return (
    <article className="flex h-full flex-col rounded-[1.75rem] border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)] transition hover:-translate-y-0.5 hover:border-[var(--border-strong)]">
      <div className="mb-4 flex flex-wrap items-start justify-between gap-3">
        <time
          dateTime={shift.dateIso}
          className="inline-flex items-center rounded-full bg-[var(--accent-soft)] px-3 py-1 text-sm font-semibold text-[var(--accent)]"
        >
          {shift.dateLabel}
        </time>
        <span className="rounded-full border border-[var(--border)] px-2.5 py-1 text-xs font-medium uppercase tracking-wide text-[var(--muted)]">
          {shift.pvzName}
        </span>
      </div>

      <h2 className="mb-2 text-xl font-semibold leading-snug text-[var(--foreground)]">
        {shift.role}
      </h2>

      <p className="mb-1 text-sm leading-6 text-[var(--foreground)]">{shift.address}</p>
      {shift.metro ? (
        <p className="mb-4 text-sm text-[var(--muted)]">{shift.metro}</p>
      ) : (
        <div className="mb-4" />
      )}

      <dl className="mt-auto grid grid-cols-1 gap-3 border-t border-[var(--border)] pt-4 text-sm sm:grid-cols-2">
        <div className="rounded-2xl bg-[var(--surface)] p-3">
          <dt className="text-[var(--muted)]">Время</dt>
          <dd className="font-medium">{shift.timeWindow}</dd>
        </div>
        <div className="rounded-2xl bg-[var(--surface)] p-3">
          <dt className="text-[var(--muted)]">Оплата</dt>
          <dd className="font-medium text-[var(--accent)]">{shift.payLabel}</dd>
        </div>
      </dl>

      <div className="mt-5 flex flex-col gap-2 sm:flex-row sm:justify-end">
        <Button type="button" variant="secondary" fullWidth className="sm:w-auto">
          Подробнее
        </Button>
        <Button type="button" fullWidth className="sm:w-auto">
          Откликнуться
        </Button>
      </div>
    </article>
  );
}
