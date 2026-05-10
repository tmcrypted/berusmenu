import type { Metadata } from "next";
import { ReviewCard } from "@/components/matching/ReviewWidgets";
import { money, shiftWindow } from "@/lib/matching-format";
import { getAssignment, getUserRating, listReviews } from "@/lib/server-matching";

export const metadata: Metadata = {
  title: "Профиль и отзывы | Беру Смену",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const userId = Number(id);
  const [ratingResult, reviewsResult] = await Promise.all([getUserRating(userId), listReviews(userId)]);
  const reviews = reviewsResult.ok ? reviewsResult.data.items : [];
  const assignmentResults = await Promise.all(reviews.map((review) => getAssignment(review.assignment_id)));

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)]">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Profile</p>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">Пользователь #{userId}</h1>
          {ratingResult.ok ? (
            <p className="mt-3 text-sm text-[var(--muted)]">Рейтинг: {ratingResult.data.avg_rating.toFixed(1)} · отзывов: {ratingResult.data.reviews_count}</p>
          ) : (
            <p className="mt-3 text-sm text-[var(--muted)]">Рейтинг пока недоступен.</p>
          )}
        </div>
        <div className="mt-6 space-y-4">
          {reviews.length === 0 ? (
            <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center text-sm text-[var(--muted)]">Отзывов пока нет.</p>
          ) : reviews.map((review, index) => {
            const assignment = assignmentResults[index]?.ok ? assignmentResults[index].data : null;
            return (
              <div key={review.id} className="space-y-3">
                {assignment ? (
                  <div className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-4 text-sm">
                    <p className="font-medium">{assignment.pvz_name_snapshot}</p>
                    <p className="mt-1 text-[var(--muted)]">{shiftWindow(assignment.shift_start_at_snapshot, assignment.shift_end_at_snapshot)} · {money(assignment.pay_amount_snapshot)}</p>
                  </div>
                ) : null}
                <ReviewCard review={review} />
              </div>
            );
          })}
        </div>
      </section>
    </div>
  );
}
