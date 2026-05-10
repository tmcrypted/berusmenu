import Link from "next/link";
import { ResponseAction } from "@/components/matching/ResponseAction";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { money, shiftWindow } from "@/lib/matching-format";
import type { User } from "@/types/auth";
import type { Vacancy } from "@/types/matching";

type Props = {
  vacancy: Vacancy;
  user?: User | null;
  hrefBase?: string;
};

export function VacancyRow({ vacancy, user, hrefBase = "/vacancies" }: Props) {
  const title = vacancy.title?.trim() || "Смена в ПВЗ";
  const isWorker = user?.role === "worker";
  const isGuest = !user;

  return (
    <article className="grid gap-4 rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)] transition hover:border-[var(--border-strong)] lg:grid-cols-[minmax(0,1fr)_auto]">
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <StatusBadge kind="vacancy" status={vacancy.status} />
          <span className="text-xs font-medium text-[var(--muted)]">PVZ #{vacancy.pvz_id}</span>
        </div>
        <Link href={`${hrefBase}/${vacancy.id}`} className="mt-3 block text-xl font-semibold tracking-tight text-[var(--foreground)] hover:text-[var(--accent)]">
          {title}
        </Link>
        <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
          {vacancy.pvz_name_snapshot} · {vacancy.pvz_address_snapshot}
        </p>
        {vacancy.description ? (
          <p className="mt-2 line-clamp-2 text-sm leading-6 text-[var(--foreground)]">{vacancy.description}</p>
        ) : null}
        <div className="mt-4 flex flex-wrap gap-2 text-sm">
          <span className="rounded-full bg-[var(--surface)] px-3 py-1 text-[var(--foreground)]">
            {shiftWindow(vacancy.shift_start_at, vacancy.shift_end_at)}
          </span>
          <span className="rounded-full bg-[var(--accent-soft)] px-3 py-1 font-semibold text-[var(--accent)]">
            {money(vacancy.pay_amount)}
          </span>
        </div>
      </div>
      <div className="flex flex-col items-stretch justify-center gap-2 lg:min-w-56">
        {(isWorker && vacancy.status === "published") || isGuest ? (
          <ResponseAction vacancyId={vacancy.id} mode={isGuest ? "guest" : "worker"} />
        ) : (
          <Link href={`${hrefBase}/${vacancy.id}`} className="inline-flex">
            <button className="inline-flex min-h-11 w-full items-center justify-center rounded-2xl border border-[var(--border)] bg-[var(--card)] px-4 py-2.5 text-sm font-semibold transition hover:bg-[var(--muted-bg)]">
              Подробнее
            </button>
          </Link>
        )}
      </div>
    </article>
  );
}
