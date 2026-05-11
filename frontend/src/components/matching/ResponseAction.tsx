"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { ApiError } from "@/lib/api";
import { createResponse } from "@/lib/matching";
import type { ResponseStatus } from "@/types/matching";

type Props = {
  vacancyId: number;
  mode: "guest" | "worker" | "complete_profile";
  existingResponseStatus?: ResponseStatus;
};

export function ResponseAction({ vacancyId, mode, existingResponseStatus }: Props) {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState(false);

  if (mode === "guest") {
    return (
      <Link href="/login" className="inline-flex">
        <Button type="button">Войти и откликнуться</Button>
      </Link>
    );
  }

  if (mode === "complete_profile") {
    return (
      <Link href="/complete-profile/worker" className="inline-flex">
        <Button type="button" variant="secondary">Завершить профиль</Button>
      </Link>
    );
  }

  async function submit() {
    setBusy(true);
    setError(null);
    try {
      await createResponse(vacancyId, message);
      setCreated(true);
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось отправить отклик.");
    } finally {
      setBusy(false);
    }
  }

  if (created || existingResponseStatus) {
    return (
      <Link href="/responses" className="inline-flex">
        <Button type="button" variant="secondary" className="border-[var(--border-strong)] bg-[var(--muted-bg)] text-[var(--muted)]">
          Вы откликнулись
        </Button>
      </Link>
    );
  }

  return (
    <div className="space-y-3">
      {!open ? (
        <Button type="button" onClick={() => setOpen(true)} className="whitespace-nowrap">
          Откликнуться
        </Button>
      ) : (
        <div className="rounded-2xl border border-[var(--border)] bg-[var(--surface)] p-3">
          <label className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]" htmlFor={`response-${vacancyId}`}>
            Сообщение владельцу, необязательно
          </label>
          <textarea
            id={`response-${vacancyId}`}
            rows={3}
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            className="mt-2 w-full rounded-xl border border-[var(--border)] bg-[var(--card)] px-3 py-2 text-sm outline-none focus:border-[var(--accent)]"
          />
          {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
          <div className="mt-3 flex flex-wrap gap-2">
            <Button type="button" disabled={busy} onClick={() => void submit()}>
              {busy ? "Отправляем..." : "Отправить отклик"}
            </Button>
            <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
              Отмена
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
