import type { Metadata } from "next";
import Link from "next/link";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { Button } from "@/components/ui/Button";
import { money, shiftWindow } from "@/lib/matching-format";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { listMyAssignments } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Мои назначения | Беру Смену",
};

export default async function Page() {
  const auth = await getServerAuthSnapshot();
  const result = auth.status === "authenticated_ready" && auth.user?.role !== "moderator" ? await listMyAssignments() : null;

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <h1 className="text-3xl font-semibold tracking-tight">Мои назначения</h1>
        {!result?.ok ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">{result?.message ?? "Раздел доступен владельцам и исполнителям."}</p>
        ) : result.data.items.length === 0 ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center text-sm text-[var(--muted)]">Назначений пока нет.</p>
        ) : (
          <div className="mt-6 space-y-4">
            {result.data.items.map((assignment) => (
              <article key={assignment.id} className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-5 shadow-[var(--shadow-soft)]">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div>
                    <StatusBadge kind="assignment" status={assignment.status} />
                    <h2 className="mt-3 text-xl font-semibold">{assignment.pvz_name_snapshot}</h2>
                    <p className="mt-1 text-sm text-[var(--muted)]">{assignment.pvz_address_snapshot}</p>
                    <p className="mt-3 text-sm">{shiftWindow(assignment.shift_start_at_snapshot, assignment.shift_end_at_snapshot)} · {money(assignment.pay_amount_snapshot)}</p>
                  </div>
                  <Link href={`/assignments/${assignment.id}`}><Button type="button" variant="secondary">Открыть</Button></Link>
                </div>
              </article>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
