import type { Metadata } from "next";
import Link from "next/link";
import { AssignmentActions, AcceptResponseButton } from "@/components/matching/MutationButtons";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { Button } from "@/components/ui/Button";
import { money, shiftWindow } from "@/lib/matching-format";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { getAssignmentForVacancy, getVacancy, listResponsesForVacancy } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Вакансия владельца | Беру Смену",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const auth = await getServerAuthSnapshot();
  const { id } = await params;
  const vacancyId = Number(id);
  const [vacancyResult, responsesResult, assignmentResult] = await Promise.all([
    getVacancy(vacancyId),
    listResponsesForVacancy(vacancyId),
    getAssignmentForVacancy(vacancyId),
  ]);

  if (auth.user?.role !== "owner") {
    return <Message text="Раздел доступен владельцу PVZ." />;
  }
  if (!vacancyResult.ok) {
    return <Message text={vacancyResult.message} />;
  }

  const vacancy = vacancyResult.data;
  const assignment = assignmentResult.ok ? assignmentResult.data : null;

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <Link href="/owner/vacancies" className="text-sm font-semibold text-[var(--accent)]">Назад к вакансиям</Link>
        <div className="mt-5 grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
          <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)]">
            <StatusBadge kind="vacancy" status={vacancy.status} />
            <h1 className="mt-4 text-3xl font-semibold">{vacancy.title?.trim() || "Смена в ПВЗ"}</h1>
            <p className="mt-2 text-sm text-[var(--muted)]">{vacancy.pvz_name_snapshot} · {vacancy.pvz_address_snapshot}</p>
            <p className="mt-4 text-sm">{shiftWindow(vacancy.shift_start_at, vacancy.shift_end_at)} · {money(vacancy.pay_amount)}</p>
            {vacancy.description ? <p className="mt-4 text-sm leading-7">{vacancy.description}</p> : null}
          </div>
          <aside className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
            <h2 className="text-lg font-semibold">Назначение</h2>
            {assignment ? (
              <div className="mt-4 space-y-3">
                <StatusBadge kind="assignment" status={assignment.status} />
                <p className="text-sm text-[var(--muted)]">Исполнитель: {assignment.worker_name_snapshot || `#${assignment.worker_id}`}</p>
                {assignment.status === "active" ? <AssignmentActions id={assignment.id} canComplete /> : null}
              </div>
            ) : (
              <p className="mt-3 text-sm text-[var(--muted)]">Исполнитель еще не выбран.</p>
            )}
          </aside>
        </div>
        <section className="mt-6 rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
          <h2 className="text-xl font-semibold">Отклики</h2>
          {!responsesResult.ok ? (
            <p className="mt-4 text-sm text-[var(--danger)]">{responsesResult.message}</p>
          ) : responsesResult.data.items.length === 0 ? (
            <p className="mt-4 text-sm text-[var(--muted)]">Откликов пока нет.</p>
          ) : (
            <div className="mt-4 space-y-3">
              {responsesResult.data.items.map((response) => (
                <div key={response.id} className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-4">
                  <div className="flex flex-wrap items-center justify-between gap-3">
                    <div>
                      <StatusBadge kind="response" status={response.status} />
                      <p className="mt-2 text-sm text-[var(--muted)]">Worker #{response.worker_id}</p>
                    </div>
                    {response.status === "pending" ? <AcceptResponseButton id={response.id} /> : null}
                  </div>
                  {response.message ? <p className="mt-3 text-sm">{response.message}</p> : null}
                </div>
              ))}
            </div>
          )}
        </section>
      </section>
    </div>
  );
}

function Message({ text }: { text: string }) {
  return (
    <div className="app-shell"><section className="mx-auto max-w-3xl px-4 py-10"><div className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-8"><p>{text}</p><Link href="/owner/vacancies" className="mt-5 inline-flex"><Button type="button">Назад</Button></Link></div></section></div>
  );
}
