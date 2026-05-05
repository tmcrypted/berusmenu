import type { ShiftCard as ShiftCardType } from "@/types/shift";

type Props = {
  shift: ShiftCardType;
};

export function ShiftCard({ shift }: Props) {
  return (
    <article className="flex flex-col rounded-2xl border border-[var(--border)] bg-[var(--card)] p-4 shadow-sm transition hover:border-[var(--accent)]/30 hover:shadow-md sm:p-5">
      <div className="mb-3 flex flex-wrap items-start justify-between gap-2">
        <time
          dateTime={shift.dateIso}
          className="inline-flex items-center rounded-lg bg-[var(--accent-soft)] px-2.5 py-1 text-sm font-semibold text-[var(--accent)]"
        >
          {shift.dateLabel}
        </time>
        <span className="text-xs font-medium uppercase tracking-wide text-[var(--muted)]">
          {shift.pvzName}
        </span>
      </div>

      <h2 className="mb-2 text-lg font-semibold leading-snug text-[var(--foreground)]">
        {shift.role}
      </h2>

      <p className="mb-1 text-sm leading-relaxed text-[var(--foreground)]">
        {shift.address}
      </p>
      {shift.metro ? (
        <p className="mb-3 text-sm text-[var(--muted)]">{shift.metro}</p>
      ) : (
        <div className="mb-3" />
      )}

      <dl className="mt-auto grid grid-cols-1 gap-2 border-t border-[var(--border)] pt-3 text-sm sm:grid-cols-2">
        <div>
          <dt className="text-[var(--muted)]">Время</dt>
          <dd className="font-medium">{shift.timeWindow}</dd>
        </div>
        <div>
          <dt className="text-[var(--muted)]">Оплата</dt>
          <dd className="font-medium text-[var(--accent)]">{shift.payLabel}</dd>
        </div>
      </dl>

      <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:justify-end">
        <button
          type="button"
          className="w-full rounded-xl border border-[var(--border)] px-4 py-2.5 text-sm font-medium transition hover:bg-[var(--muted-bg)] sm:w-auto"
        >
          Подробнее
        </button>
        <button
          type="button"
          className="w-full rounded-xl bg-[var(--accent)] px-4 py-2.5 text-sm font-medium text-white transition hover:bg-[var(--accent-hover)] sm:w-auto"
        >
          Откликнуться
        </button>
      </div>
    </article>
  );
}
