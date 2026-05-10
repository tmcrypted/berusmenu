"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { ApiError } from "@/lib/api";
import { createResponse } from "@/lib/matching";

type Props = {
  vacancyId: number;
  mode: "guest" | "worker";
};

export function ResponseAction({ vacancyId, mode }: Props) {
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

  if (created) {
    return (
      <Link href="/responses" className="inline-flex text-sm font-semibold text-[var(--accent)]">
        Отклик отправлен. Перейти к моим откликам
      </Link>
    );
  }

  return (
    <div className="space-y-3">
      {!open ? (
        <Button type="button" onClick={() => setOpen(true)}>
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
