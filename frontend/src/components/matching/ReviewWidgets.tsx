"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { ApiError } from "@/lib/api";
import { createReview, createReviewReply } from "@/lib/matching";
import { dateOnly, dateTime, money } from "@/lib/matching-format";
import { shortDisplayName } from "@/lib/profile";
import type { PublicUser } from "@/types/auth";
import type { Assignment, Review, ReviewReply } from "@/types/matching";

export function ReviewForm({ assignmentId }: { assignmentId: number }) {
  const router = useRouter();
  const [rating, setRating] = useState(5);
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function submit() {
    setBusy(true);
    setError(null);
    try {
      await createReview(assignmentId, rating, body);
      setBody("");
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось сохранить отзыв.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-5">
      <h2 className="text-lg font-semibold">Оставить отзыв</h2>
      <div className="mt-4 grid gap-3 md:grid-cols-[10rem_1fr_auto]">
        <StarRating value={rating} onChange={setRating} />
        <input value={body} onChange={(event) => setBody(event.target.value)} placeholder="Комментарий, необязательно" className="min-h-11 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3" />
        <Button type="button" disabled={busy} onClick={() => void submit()}>{busy ? "Сохраняем..." : "Отправить"}</Button>
      </div>
      {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
    </div>
  );
}

export function ReviewCard({
  review,
  initialReplies,
  author,
  assignment,
}: {
  review: Review;
  initialReplies?: ReviewReply[];
  author?: PublicUser | null;
  assignment?: Assignment | null;
}) {
  const [open, setOpen] = useState(false);
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [replies, setReplies] = useState(initialReplies ?? []);
  const [error, setError] = useState<string | null>(null);
  const authorName = shortDisplayName(author, `Пользователь #${review.author_user_id}`);

  async function reply() {
    setBusy(true);
    setError(null);
    try {
      const target = review.author_user_id;
      const created = await createReviewReply(review.id, target, body);
      setReplies((current) => [...current, created]);
      setBody("");
      setOpen(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось ответить на отзыв.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <article className="rounded-2xl border border-[var(--border)] bg-[var(--card)] px-4 py-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <ReadonlyStars value={review.rating} />
            <span className="text-xs text-[var(--muted)]">
              от{" "}
              <Link href={`/profile/${review.author_user_id}`} className="font-medium hover:text-[var(--accent)]">
                {authorName}
              </Link>
            </span>
          </div>
          {assignment ? (
            <p className="mt-1 truncate text-xs text-[var(--muted)]">
              {assignment.pvz_name_snapshot}{assignmentLocation(assignment) ? `, ${assignmentLocation(assignment)}` : ""}, {dateOnly(assignment.shift_start_at_snapshot)}, {money(assignment.pay_amount_snapshot)}
            </p>
          ) : (
            <p className="mt-1 text-xs text-[var(--muted)]">Смена #{review.assignment_id}</p>
          )}
        </div>
        <span className="shrink-0 text-xs text-[var(--muted)]">{dateTime(review.created_at)}</span>
      </div>
      <p className="mt-2 text-sm leading-5 text-[var(--foreground)]">{review.body?.trim() || "Без комментария"}</p>
      <button
        type="button"
        onClick={() => setOpen((value) => !value)}
        className="mt-2 inline-flex h-7 items-center gap-1 rounded-full border border-[var(--border)] bg-[var(--surface)] px-2 text-xs font-semibold text-[var(--muted)] transition hover:bg-[var(--muted-bg)]"
        aria-expanded={open}
      >
        <span aria-hidden>{open ? "▲" : "▼"}</span>
        {replies.length}
      </button>
      {open ? (
        <div className="mt-2 space-y-2">
          {replies.map((replyItem) => (
            <div key={replyItem.id} className="rounded-xl bg-[var(--surface)] px-3 py-2 text-sm">
              {replyItem.body}
            </div>
          ))}
        </div>
      ) : null}
      <div className="mt-3 flex flex-col gap-2 sm:flex-row">
        <input value={body} onChange={(event) => setBody(event.target.value)} placeholder="Ответить" className="min-h-9 flex-1 rounded-xl border border-[var(--border)] bg-[var(--surface)] px-3 text-sm outline-none focus:border-[var(--accent)]" />
        <Button type="button" disabled={busy || !body.trim()} onClick={() => void reply()} className="min-h-9 rounded-xl px-3 py-1.5 text-xs">
          {busy ? "..." : "Ответить"}
        </Button>
      </div>
      {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
    </article>
  );
}

function assignmentLocation(assignment: Assignment) {
  const city = assignment.pvz_city_snapshot?.trim();
  const address = assignment.pvz_address_snapshot.trim();
  if (!city) {
    return address;
  }
  if (!address || address.toLowerCase().includes(city.toLowerCase())) {
    return address || city;
  }
  return `${city}, ${address}`;
}

function StarRating({ value, onChange }: { value: number; onChange: (value: number) => void }) {
  return (
    <div className="flex min-h-11 items-center gap-1 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3" role="radiogroup" aria-label="Оценка">
      {[1, 2, 3, 4, 5].map((star) => (
        <button
          key={star}
          type="button"
          role="radio"
          aria-checked={value === star}
          aria-label={`${star} из 5`}
          onClick={() => onChange(star)}
          className={star <= value ? "text-xl leading-none text-[var(--accent)]" : "text-xl leading-none text-[var(--muted)]"}
        >
          ★
        </button>
      ))}
    </div>
  );
}

function ReadonlyStars({ value }: { value: number }) {
  return (
    <p className="text-sm font-semibold leading-none text-[var(--accent)]" aria-label={`Оценка ${value} из 5`}>
      {[1, 2, 3, 4, 5].map((star) => (star <= value ? "★" : "☆")).join("")}
    </p>
  );
}
