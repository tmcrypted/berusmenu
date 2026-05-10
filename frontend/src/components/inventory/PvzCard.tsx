import Link from "next/link";
import { PvzStatusBadge } from "@/components/inventory/PvzStatusBadge";
import { Button } from "@/components/ui/Button";
import type { Pvz } from "@/types/inventory";

type Props = {
  pvz: Pvz;
};

export function PvzCard({ pvz }: Props) {
  const canEdit = pvz.status !== "blocked";

  return (
    <article className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)]">
      <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
        <div className="min-w-0">
          <h2 className="truncate text-lg font-semibold text-[var(--foreground)]">
            {pvz.company_name}
          </h2>
          <p className="mt-1 text-sm text-[var(--muted)]">{pvz.city}, {pvz.address}</p>
          <p className="mt-1 text-sm text-[var(--muted)]">{pvz.contact_phone}</p>
        </div>
        <PvzStatusBadge status={pvz.status} />
      </div>

      {pvz.description ? (
        <p className="mt-4 text-sm leading-6 text-[var(--foreground)]">{pvz.description}</p>
      ) : null}

      {pvz.rejection_reason ? (
        <div className="mt-4 rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-900">
          Причина отклонения: {pvz.rejection_reason}
        </div>
      ) : null}

      {pvz.status === "approved" ? (
        <p className="mt-4 text-sm text-[var(--muted)]">
          ПВЗ одобрен. Создание вакансий появится после подключения matching-flow.
        </p>
      ) : pvz.status === "blocked" ? (
        <p className="mt-4 text-sm text-[var(--muted)]">
          ПВЗ заблокирован, редактирование недоступно.
        </p>
      ) : null}

      <div className="mt-5 flex flex-wrap gap-3">
        {pvz.status === "approved" ? (
          <Link href={`/owner/vacancies/new?pvz_id=${pvz.id}`}>
            <Button type="button" className="rounded-xl">
              Создать вакансию
            </Button>
          </Link>
        ) : null}
        {canEdit ? (
          <Link href={`/owner/pvz/${pvz.id}/edit`}>
            <Button type="button" variant="secondary" className="rounded-xl">
              Редактировать
            </Button>
          </Link>
        ) : null}
      </div>
    </article>
  );
}
