"use client";

import { useEffect, useState } from "react";
import { InventoryRoleGuard } from "@/components/inventory/InventoryRoleGuard";
import { PvzStatusBadge } from "@/components/inventory/PvzStatusBadge";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { Field } from "@/components/ui/Field";
import { Input } from "@/components/ui/Input";
import { Spinner } from "@/components/ui/Spinner";
import { ApiError } from "@/lib/api";
import {
  approvePvz,
  blockPvz,
  blockUserAndDeletePvzRequests,
  listPendingPvz,
  rejectPvz,
} from "@/lib/inventory";
import type { ModerationDecisionInput, Pvz } from "@/types/inventory";

type Decision = "approve" | "reject" | "block" | "block_user";

export function ModeratorPvzQueue() {
  const [items, setItems] = useState<Pvz[]>([]);
  const [selected, setSelected] = useState<Pvz | null>(null);
  const [nextPageToken, setNextPageToken] = useState<string | undefined>();
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [actionLoading, setActionLoading] = useState<Decision | null>(null);
  const [reason, setReason] = useState("");
  const [blockDurationMinutes, setBlockDurationMinutes] = useState(60);
  const [comment, setComment] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [decisionError, setDecisionError] = useState<string | null>(null);
  const [decisionMessage, setDecisionMessage] = useState<string | null>(null);

  async function load(pageToken?: string) {
    if (pageToken) {
      setLoadingMore(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      const response = await listPendingPvz(pageToken);
      setItems((current) => (pageToken ? [...current, ...response.items] : response.items));
      setNextPageToken(response.next_page_token);
      if (!pageToken) {
        setSelected(response.items[0] ?? null);
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось загрузить очередь модерации.");
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function decide(decision: Decision) {
    if (!selected) {
      return;
    }
    if ((decision === "reject" || decision === "block_user") && !reason.trim()) {
      setDecisionError("Для отклонения или блокировки пользователя нужна причина.");
      return;
    }
    if (decision === "block_user" && (!Number.isFinite(blockDurationMinutes) || blockDurationMinutes < 1)) {
      setDecisionError("Срок блокировки должен быть больше 0 минут.");
      return;
    }

    setActionLoading(decision);
    setDecisionError(null);
    setDecisionMessage(null);
    const input: ModerationDecisionInput = { reason, comment };

    try {
      if (decision === "approve") {
        await approvePvz(selected.id, input);
      } else if (decision === "reject") {
        await rejectPvz(selected.id, input);
      } else if (decision === "block_user") {
        const result = await blockUserAndDeletePvzRequests(selected.owner_id, {
          reason,
          block_duration_minutes: blockDurationMinutes,
          comment,
        });
        const remaining = items.filter((item) => item.owner_id !== selected.owner_id);
        setItems(remaining);
        setSelected(remaining[0] ?? null);
        setReason("");
        setComment("");
        setDecisionMessage(
          `Пользователь заблокирован до ${dateLabel(result.blocked_until)}. Удалено заявок: ${result.deleted_count}, заблокировано ПВЗ: ${result.blocked_count}.`,
        );
        return;
      } else {
        await blockPvz(selected.id, input);
      }

      const remaining = items.filter((item) => item.id !== selected.id);
      setItems(remaining);
      setSelected(remaining[0] ?? null);
      setReason("");
      setComment("");
    } catch (err) {
      setDecisionError(err instanceof ApiError ? err.message : "Не удалось применить решение.");
    } finally {
      setActionLoading(null);
    }
  }

  return (
    <InventoryRoleGuard role="moderator">
      <div className="app-shell">
        <section className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
              Moderation
            </p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight text-[var(--foreground)]">
              Модерация ПВЗ
            </h1>
            <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)]">
              Проверяйте заявки владельцев. Аккаунт модератора назначается только через базу данных.
            </p>
          </div>

          {error ? (
            <div className="mt-6">
              <Alert tone="error">{error}</Alert>
            </div>
          ) : null}

          {loading ? (
            <div className="mt-16 flex justify-center">
              <Spinner />
            </div>
          ) : items.length === 0 ? (
            <div className="mt-8 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center shadow-[var(--shadow-soft)]">
              <h2 className="text-xl font-semibold text-[var(--foreground)]">Очередь пуста</h2>
              <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-[var(--muted)]">
                Сейчас нет ПВЗ в статусе pending.
              </p>
            </div>
          ) : (
            <div className="mt-8 grid gap-6 lg:grid-cols-[minmax(18rem,0.75fr)_minmax(0,1.25fr)]">
              <div className="space-y-3">
                {items.map((pvz) => (
                  <button
                    key={pvz.id}
                    type="button"
                    onClick={() => {
                      setSelected(pvz);
                      setDecisionError(null);
                      setDecisionMessage(null);
                      setReason("");
                      setComment("");
                    }}
                    className={`w-full rounded-2xl border p-4 text-left transition ${
                      selected?.id === pvz.id
                        ? "border-[var(--accent)] bg-[var(--accent-soft)]"
                        : "border-[var(--border)] bg-[var(--card)] hover:bg-[var(--muted-bg)]"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-3">
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold text-[var(--foreground)]">
                          {pvz.company_name}
                        </p>
                        <p className="mt-1 text-xs text-[var(--muted)]">{pvz.city}</p>
                      </div>
                      <PvzStatusBadge status={pvz.status} />
                    </div>
                  </button>
                ))}
                {nextPageToken ? (
                  <Button
                    type="button"
                    variant="secondary"
                    fullWidth
                    disabled={loadingMore}
                    onClick={() => void load(nextPageToken)}
                  >
                    {loadingMore ? "Загружаем..." : "Показать еще"}
                  </Button>
                ) : null}
              </div>

              {selected ? (
                <div className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)]">
                  <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                    <div>
                      <h2 className="text-2xl font-semibold tracking-tight text-[var(--foreground)]">
                        {selected.company_name}
                      </h2>
                      <p className="mt-1 text-sm text-[var(--muted)]">
                        {selected.city}, {selected.address}
                      </p>
                    </div>
                    <PvzStatusBadge status={selected.status} />
                  </div>

                  <dl className="mt-6 grid gap-4 text-sm sm:grid-cols-2">
                    <Info label="Телефон" value={selected.contact_phone} />
                    <Info label="Telegram" value={selected.contact_telegram || "-"} />
                    <Info label="Координаты" value={coordinateLabel(selected)} />
                    <Info label="Создан" value={dateLabel(selected.created_at)} />
                  </dl>

                  {selected.description ? (
                    <div className="mt-6 rounded-xl border border-[var(--border)] bg-[var(--surface)] p-4 text-sm leading-6 text-[var(--foreground)]">
                      {selected.description}
                    </div>
                  ) : null}

                  <div className="mt-6 space-y-4">
                    {decisionError ? <Alert tone="error">{decisionError}</Alert> : null}
                    {decisionMessage ? <Alert tone="success">{decisionMessage}</Alert> : null}
                    <Field label="Причина решения" htmlFor="moderation-reason" hint="Обязательно для reject и блокировки пользователя">
                      <Input
                        id="moderation-reason"
                        value={reason}
                        onChange={(event) => setReason(event.target.value)}
                      />
                    </Field>
                    <Field label="Срок блокировки, минут" htmlFor="block-duration" hint="Используется только для блокировки пользователя">
                      <Input
                        id="block-duration"
                        type="number"
                        min={1}
                        value={blockDurationMinutes}
                        onChange={(event) => setBlockDurationMinutes(Number(event.target.value))}
                      />
                    </Field>
                    <Field label="Комментарий" htmlFor="moderation-comment" hint="Необязательно">
                      <textarea
                        id="moderation-comment"
                        rows={3}
                        value={comment}
                        onChange={(event) => setComment(event.target.value)}
                        className="w-full rounded-2xl border border-[var(--border)] bg-[var(--surface)] px-3.5 py-2.5 text-sm text-[var(--foreground)] outline-none transition focus:border-[var(--accent)] focus:ring-4 focus:ring-[var(--accent-soft)]"
                      />
                    </Field>
                    <div className="grid gap-3 sm:grid-cols-4">
                      <Button type="button" disabled={Boolean(actionLoading)} onClick={() => void decide("approve")}>
                        {actionLoading === "approve" ? "..." : "Одобрить"}
                      </Button>
                      <Button type="button" variant="secondary" disabled={Boolean(actionLoading)} onClick={() => void decide("reject")}>
                        {actionLoading === "reject" ? "..." : "Отклонить"}
                      </Button>
                      <Button type="button" variant="secondary" disabled={Boolean(actionLoading)} onClick={() => void decide("block")}>
                        {actionLoading === "block" ? "..." : "Блок ПВЗ"}
                      </Button>
                      <Button
                        type="button"
                        variant="secondary"
                        disabled={Boolean(actionLoading)}
                        className="border-red-200 text-red-700 hover:bg-red-50"
                        onClick={() => void decide("block_user")}
                      >
                        {actionLoading === "block_user" ? "..." : "Блок owner"}
                      </Button>
                    </div>
                  </div>
                </div>
              ) : null}
            </div>
          )}
        </section>
      </div>
    </InventoryRoleGuard>
  );
}

function Info({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <dt className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">{label}</dt>
      <dd className="mt-1 text-[var(--foreground)]">{value}</dd>
    </div>
  );
}

function coordinateLabel(pvz: Pvz) {
  if (typeof pvz.lat === "number" && typeof pvz.lng === "number") {
    return `${pvz.lat}, ${pvz.lng}`;
  }
  return "-";
}

function dateLabel(value: string) {
  if (!value) {
    return "-";
  }
  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(new Date(value));
}
