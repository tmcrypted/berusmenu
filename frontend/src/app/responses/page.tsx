import type { Metadata } from "next";
import Link from "next/link";
import { ChatPlaceholder } from "@/components/matching/ChatPlaceholder";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { WithdrawResponseButton } from "@/components/matching/MutationButtons";
import { Button } from "@/components/ui/Button";
import { dateTime } from "@/lib/matching-format";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { listMyResponses } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Мои отклики | Беру Смену",
};

export default async function Page() {
  const auth = await getServerAuthSnapshot();
  const result = auth.user?.role === "worker" ? await listMyResponses() : null;

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-semibold tracking-tight">Мои отклики</h1>
        {auth.user?.role !== "worker" ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Раздел доступен исполнителю.</p>
        ) : !result?.ok ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{result?.message ?? "Не удалось загрузить отклики."}</p>
        ) : result.data.items.length === 0 ? (
          <div className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center">
            <p className="text-sm text-[var(--muted)]">Вы пока не откликались.</p>
            <Link href="/vacancies" className="mt-5 inline-flex"><Button type="button">Открыть ленту</Button></Link>
          </div>
        ) : (
          <div className="mt-6 space-y-4">
            {result.data.items.map((response) => (
              <article key={response.id} className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)]">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <StatusBadge kind="response" status={response.status} />
                    <h2 className="mt-3 text-xl font-semibold">Отклик на vacancy #{response.vacancy_id}</h2>
                    <p className="mt-1 text-sm text-[var(--muted)]">Создан: {dateTime(response.created_at)}</p>
                  </div>
                  {response.status === "pending" ? <WithdrawResponseButton id={response.id} /> : null}
                </div>
                {response.message ? <p className="mt-4 text-sm leading-6">{response.message}</p> : null}
                <div className="mt-5"><ChatPlaceholder response={response} /></div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
