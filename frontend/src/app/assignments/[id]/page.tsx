import type { Metadata } from "next";
import Link from "next/link";
import { AssignmentActions } from "@/components/matching/MutationButtons";
import { ChatPlaceholder } from "@/components/matching/ChatPlaceholder";
import { ReviewForm, ReviewCard } from "@/components/matching/ReviewWidgets";
import { StatusBadge } from "@/components/matching/StatusBadge";
import { Button } from "@/components/ui/Button";
import { money, shiftWindow } from "@/lib/matching-format";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { getAssignment, listReviews } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Назначение | Беру Смену",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const auth = await getServerAuthSnapshot();
  const { id } = await params;
  const assignmentResult = await getAssignment(Number(id));

  if (!assignmentResult.ok) {
    return <Message text={assignmentResult.message} />;
  }

  const assignment = assignmentResult.data;
  const canComplete = auth.user?.id === assignment.owner_id;
  const reviewsResult = await listReviews();
  const assignmentReviews = reviewsResult.ok
    ? reviewsResult.data.items.filter((review) => review.assignment_id === assignment.id)
    : [];

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <Link href="/assignments" className="text-sm font-semibold text-[var(--accent)]">Назад к назначениям</Link>
        <div className="mt-5 grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem]">
          <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)]">
            <StatusBadge kind="assignment" status={assignment.status} />
            <h1 className="mt-4 text-3xl font-semibold">{assignment.pvz_name_snapshot}</h1>
            <p className="mt-2 text-sm text-[var(--muted)]">{assignment.pvz_address_snapshot}</p>
            <p className="mt-4 text-sm">{shiftWindow(assignment.shift_start_at_snapshot, assignment.shift_end_at_snapshot)} · {money(assignment.pay_amount_snapshot)}</p>
            <dl className="mt-6 grid gap-3 sm:grid-cols-2">
              <Info label="Владелец" value={assignment.owner_name_snapshot || `#${assignment.owner_id}`} />
              <Info label="Исполнитель" value={assignment.worker_name_snapshot || `#${assignment.worker_id}`} />
            </dl>
          </div>
          <aside className="space-y-4 rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
            <h2 className="text-lg font-semibold">Действия</h2>
            {assignment.status === "active" ? <AssignmentActions id={assignment.id} canComplete={canComplete} /> : <p className="text-sm text-[var(--muted)]">Смена финализирована.</p>}
          </aside>
        </div>
        <div className="mt-6"><ChatPlaceholder assignment={assignment} /></div>
        {assignment.status !== "active" ? (
          <div className="mt-6 space-y-4">
            <ReviewForm assignmentId={assignment.id} />
            {assignmentReviews.map((review) => (
              <ReviewCard key={review.id} review={review} />
            ))}
          </div>
        ) : null}
      </section>
    </div>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return <div className="rounded-2xl bg-[var(--surface)] p-3"><dt className="text-xs uppercase tracking-[0.12em] text-[var(--muted)]">{label}</dt><dd className="mt-1 font-medium">{value}</dd></div>;
}

function Message({ text }: { text: string }) {
  return <div className="app-shell"><section className="mx-auto max-w-3xl px-4 py-10"><div className="rounded-[2rem] border border-[var(--border)] bg-[var(--card)] p-8"><p>{text}</p><Link href="/assignments" className="mt-5 inline-flex"><Button type="button">Назад</Button></Link></div></section></div>;
}
