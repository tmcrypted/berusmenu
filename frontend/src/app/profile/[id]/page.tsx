import type { Metadata } from "next";
import { ReviewCard } from "@/components/matching/ReviewWidgets";
import { UserAvatar } from "@/components/profile/UserAvatar";
import { ageLabel, displayName, roleLabel } from "@/lib/profile";
import { getAssignment, getUserRating, listReviews } from "@/lib/server-matching";
import { getPublicUser } from "@/lib/server-users";
import type { PublicUser } from "@/types/auth";

export const metadata: Metadata = {
  title: "Профиль и отзывы | Беру Смену",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const userId = Number(id);
  const [userResult, ratingResult, reviewsResult] = await Promise.all([
    getPublicUser(userId),
    getUserRating(userId),
    listReviews(userId),
  ]);
  const user = userResult.ok ? userResult.data : null;
  const reviews = reviewsResult.ok ? reviewsResult.data.items : [];
  const [assignmentResults, authorResults] = await Promise.all([
    Promise.all(reviews.map((review) => getAssignment(review.assignment_id))),
    Promise.all([...new Set(reviews.map((review) => review.author_user_id))].map((authorId) => getPublicUser(authorId))),
  ]);
  const authorIds = [...new Set(reviews.map((review) => review.author_user_id))];
  const authors = new Map<number, PublicUser>();
  authorResults.forEach((result, index) => {
    if (result.ok) {
      authors.set(authorIds[index], result.data);
    }
  });
  const userName = displayName(user, `Пользователь #${userId}`);

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <div className="flex flex-col gap-5 sm:flex-row sm:items-center">
            <UserAvatar name={userName} src={user?.avatar_url} size="lg" />
            <div className="min-w-0">
              <h1 className="text-3xl font-semibold tracking-tight">{userName}</h1>
              <p className="mt-2 text-sm text-[var(--muted)]">
                {user ? roleLabel(user.role) : "Профиль недоступен"} · {ageLabel(user?.birth_date)}
              </p>
              {ratingResult.ok ? (
                <p className="mt-3 text-sm text-[var(--muted)]">
                  Рейтинг: {ratingResult.data.avg_rating.toFixed(1)} · отзывов: {ratingResult.data.reviews_count}
                </p>
              ) : (
                <p className="mt-3 text-sm text-[var(--muted)]">Рейтинг пока недоступен.</p>
              )}
              {user?.telegram_username || user?.max_username ? (
                <div className="mt-3 flex flex-wrap gap-2 text-sm">
                  {user.telegram_username ? <Contact label="Telegram" value={user.telegram_username} /> : null}
                  {user.max_username ? <Contact label="Max" value={user.max_username} /> : null}
                </div>
              ) : null}
            </div>
          </div>
        </div>
        <div className="mt-6 space-y-4">
          {reviews.length === 0 ? (
            <p className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center text-sm text-[var(--muted)]">
              Отзывов пока нет.
            </p>
          ) : (
            reviews.map((review, index) => {
              const assignment = assignmentResults[index]?.ok ? assignmentResults[index].data : null;
              return (
                <ReviewCard key={review.id} review={review} assignment={assignment} author={authors.get(review.author_user_id)} />
              );
            })
          )}
        </div>
      </section>
    </div>
  );
}

function Contact({ label, value }: { label: string; value: string }) {
  return (
    <span className="rounded-full bg-[var(--surface)] px-3 py-1 text-[var(--muted)]">
      {label}: @{value.replace(/^@+/, "")}
    </span>
  );
}
