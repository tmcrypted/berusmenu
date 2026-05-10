"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "@/components/ui/Button";
import { ApiError } from "@/lib/api";
import {
  acceptResponse,
  cancelAssignment,
  completeAssignment,
  moderationAcceptResponse,
  moderationCancelAssignment,
  moderationCompleteAssignment,
  withdrawResponse,
} from "@/lib/matching";

type BasicProps = {
  id: number;
};

export function WithdrawResponseButton({ id }: BasicProps) {
  const router = useRouter();
  const [confirming, setConfirming] = useState(false);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function run() {
    setBusy(true);
    setError(null);
    try {
      await withdrawResponse(id);
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось отозвать отклик.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="space-y-2">
      {!confirming ? (
        <Button type="button" variant="secondary" onClick={() => setConfirming(true)}>
          Отозвать
        </Button>
      ) : (
        <div className="rounded-2xl border border-[var(--danger)]/30 bg-[var(--surface)] p-3">
          <p className="text-sm font-medium">Отклик нельзя будет отправить повторно.</p>
          <div className="mt-3 flex gap-2">
            <Button type="button" disabled={busy} onClick={() => void run()}>
              {busy ? "Отзываем..." : "Да, отозвать"}
            </Button>
            <Button type="button" variant="secondary" onClick={() => setConfirming(false)}>
              Не отзывать
            </Button>
          </div>
        </div>
      )}
      {error ? <p className="text-sm text-[var(--danger)]">{error}</p> : null}
    </div>
  );
}

export function AcceptResponseButton({ id }: BasicProps) {
  return <AsyncButton label="Принять отклик" pendingLabel="Принимаем..." action={() => acceptResponse(id)} />;
}

export function AssignmentActions({ id, canComplete }: BasicProps & { canComplete: boolean }) {
  const [reason, setReason] = useState("");
  return (
    <div className="space-y-3">
      <input
        value={reason}
        onChange={(event) => setReason(event.target.value)}
        placeholder="Причина отмены, необязательно"
        className="min-h-11 w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm outline-none"
      />
      <div className="flex flex-wrap gap-2">
        {canComplete ? <AsyncButton label="Завершить смену" pendingLabel="Завершаем..." action={() => completeAssignment(id)} /> : null}
        <AsyncButton label="Отменить смену" pendingLabel="Отменяем..." variant="secondary" action={() => cancelAssignment(id, reason)} />
      </div>
    </div>
  );
}

export function ModeratorResponseAction({ id }: BasicProps) {
  return <ReasonButton label="Принять спорный отклик" action={(reason) => moderationAcceptResponse(id, reason)} />;
}

export function ModeratorAssignmentActions({ id }: BasicProps) {
  const [cancelReason, setCancelReason] = useState("");
  return (
    <div className="space-y-2">
      <input
        value={cancelReason}
        onChange={(event) => setCancelReason(event.target.value)}
        placeholder="Причина отмены для assignment, необязательно"
        className="min-h-11 w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm outline-none"
      />
      <ReasonButton label="Отменить assignment" action={(reason) => moderationCancelAssignment(id, reason, cancelReason)} />
      <ReasonButton label="Завершить assignment" action={(reason) => moderationCompleteAssignment(id, reason)} />
    </div>
  );
}

function AsyncButton({
  label,
  pendingLabel,
  action,
  variant,
}: {
  label: string;
  pendingLabel: string;
  action: () => Promise<unknown>;
  variant?: "primary" | "secondary";
}) {
  const router = useRouter();
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function run() {
    setBusy(true);
    setError(null);
    try {
      await action();
      router.refresh();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Действие не выполнено.");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div>
      <Button type="button" variant={variant} disabled={busy} onClick={() => void run()}>
        {busy ? pendingLabel : label}
      </Button>
      {error ? <p className="mt-2 text-sm text-[var(--danger)]">{error}</p> : null}
    </div>
  );
}

function ReasonButton({ label, action }: { label: string; action: (reason: string) => Promise<unknown> }) {
  const [reason, setReason] = useState("");
  return (
    <div className="space-y-2">
      <input
        value={reason}
        onChange={(event) => setReason(event.target.value)}
        placeholder="Причина модератора"
        className="min-h-11 w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm outline-none"
      />
      <AsyncButton label={label} pendingLabel="Выполняем..." action={() => action(reason)} />
    </div>
  );
}
