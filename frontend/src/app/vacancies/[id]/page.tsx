import type { Metadata } from "next";
import Link from "next/link";
import { ChatPlaceholder } from "@/components/matching/ChatPlaceholder";
import { ResponseAction } from "@/components/matching/ResponseAction";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { Button } from "@/components/ui/Button";
import { money, shiftWindow } from "@/lib/matching-format";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { getVacancy } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Вакансия | Беру Смену",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const auth = await getServerAuthSnapshot();
  const result = await getVacancy(Number(id));

  if (!result.ok) {
    return <Message title="Вакансия недоступна" message={result.message} />;
  }

  const vacancy = result.data;
  const isWorker = auth.user?.role === "worker";

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <Link href="/vacancies" className="text-sm font-semibold text-[var(--accent)]">Назад к ленте</Link>
        <div className="mt-5 rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <StatusBadge kind="vacancy" status={vacancy.status} />
          <h1 className="mt-4 text-3xl font-semibold tracking-tight">{vacancy.title?.trim() || "Смена в ПВЗ"}</h1>
          <p className="mt-3 text-sm leading-6 text-[var(--muted)]">{vacancy.pvz_name_snapshot} · {vacancy.pvz_address_snapshot}</p>
          <div className="mt-5 flex flex-wrap gap-2">
            <span className="rounded-full bg-[var(--surface)] px-3 py-1 text-sm">{shiftWindow(vacancy.shift_start_at, vacancy.shift_end_at)}</span>
            <span className="rounded-full bg-[var(--accent-soft)] px-3 py-1 text-sm font-semibold text-[var(--accent)]">{money(vacancy.pay_amount)}</span>
          </div>
          {vacancy.description ? <p className="mt-6 text-sm leading-7">{vacancy.description}</p> : null}
          <div className="mt-8">
            {isWorker && vacancy.status === "published" ? (
              <ResponseAction vacancyId={vacancy.id} mode="worker" />
            ) : auth.status === "anonymous" ? (
              <ResponseAction vacancyId={vacancy.id} mode="guest" />
            ) : null}
          </div>
        </div>
        {isWorker ? <div className="mt-6"><ChatPlaceholder /></div> : null}
      </section>
    </div>
  );
}

function Message({ title, message }: { title: string; message: string }) {
  return (
    <div className="app-shell">
      <section className="mx-auto max-w-3xl px-4 py-10">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-8">
          <h1 className="text-2xl font-semibold">{title}</h1>
          <p className="mt-2 text-sm text-[var(--muted)]">{message}</p>
          <Link href="/vacancies" className="mt-5 inline-flex"><Button type="button">К ленте</Button></Link>
        </div>
      </section>
    </div>
  );
}
