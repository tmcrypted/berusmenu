"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { ApiError } from "@/lib/api";
import { createReview, createReviewReply } from "@/lib/matching";
import type { Review, ReviewReply } from "@/types/matching";

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
      <div className="mt-4 grid gap-3 md:grid-cols-[8rem_1fr_auto]">
        <select value={rating} onChange={(event) => setRating(Number(event.target.value))} className="min-h-11 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3">
          {[5, 4, 3, 2, 1].map((value) => (
            <option key={value} value={value}>{value}</option>
          ))}
        </select>
        <input value={body} onChange={(event) => setBody(event.target.value)} placeholder="Комментарий, необязательно" className="min-h-11 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3" />
        <Button type="button" disabled={busy} onClick={() => void submit()}>{busy ? "Сохраняем..." : "Отправить"}</Button>
      </div>
      {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
    </div>
  );
}

export function ReviewCard({ review, initialReplies }: { review: Review; initialReplies?: ReviewReply[] }) {
  const [open, setOpen] = useState(false);
  const [body, setBody] = useState("");
  const [busy, setBusy] = useState(false);
  const [replies, setReplies] = useState(initialReplies ?? []);
  const [error, setError] = useState<string | null>(null);

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
    <article className="rounded-[1.5rem] border border-[var(--border)] bg-[var(--card)] p-5">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="font-semibold">Оценка: {review.rating}/5</p>
        <span className="text-xs text-[var(--muted)]">Assignment #{review.assignment_id}</span>
      </div>
      <p className="mt-3 text-sm leading-6 text-[var(--foreground)]">{review.body?.trim() || "Без комментария"}</p>
      <div className="mt-4 flex flex-wrap gap-2">
        <Button type="button" variant="secondary" onClick={() => setOpen((value) => !value)}>
          {open ? "Скрыть ответы" : `Показать ответы (${replies.length})`}
        </Button>
      </div>
      {open ? (
        <div className="mt-4 space-y-3">
          {replies.map((replyItem) => (
            <div key={replyItem.id} className="rounded-2xl bg-[var(--surface)] p-3 text-sm">
              {replyItem.body}
            </div>
          ))}
        </div>
      ) : null}
      <div className="mt-4 flex flex-col gap-2 sm:flex-row">
        <input value={body} onChange={(event) => setBody(event.target.value)} placeholder="Ответить на отзыв" className="min-h-11 flex-1 rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3" />
        <Button type="button" disabled={busy || !body.trim()} onClick={() => void reply()}>{busy ? "Отвечаем..." : "Ответить"}</Button>
      </div>
      {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
    </article>
  );
}
