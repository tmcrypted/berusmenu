"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { InventoryRoleGuard } from "@/components/inventory/InventoryRoleGuard";
import { PvzCard } from "@/components/inventory/PvzCard";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { Spinner } from "@/components/ui/Spinner";
import { ApiError } from "@/lib/api";
import { listOwnerPvz } from "@/lib/inventory";
import type { Pvz } from "@/types/inventory";

export function OwnerPvzPage() {
  const [items, setItems] = useState<Pvz[]>([]);
  const [nextPageToken, setNextPageToken] = useState<string | undefined>();
  const [showSubmitted, setShowSubmitted] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function load(pageToken?: string) {
    if (pageToken) {
      setLoadingMore(true);
    } else {
      setLoading(true);
    }
    setError(null);
    try {
      const response = await listOwnerPvz(pageToken);
      setItems((current) => (pageToken ? [...current, ...response.items] : response.items));
      setNextPageToken(response.next_page_token);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "Не удалось загрузить ПВЗ.");
    } finally {
      setLoading(false);
      setLoadingMore(false);
    }
  }

  useEffect(() => {
    setShowSubmitted(new URLSearchParams(window.location.search).has("submitted"));
    void load();
  }, []);

  return (
    <InventoryRoleGuard role="owner">
      <div className="app-shell">
        <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
                Inventory
              </p>
              <h1 className="mt-2 text-3xl font-semibold tracking-tight text-[var(--foreground)]">
                Мои ПВЗ
              </h1>
              <p className="mt-2 max-w-2xl text-sm leading-6 text-[var(--muted)]">
                Создавайте ПВЗ и отслеживайте статус модерации. Вакансии можно будет создавать после одобрения ПВЗ.
              </p>
            </div>
            <Link href="/owner/pvz/new">
              <Button type="button">Новый ПВЗ</Button>
            </Link>
          </div>

          {showSubmitted ? (
            <div className="mt-6">
              <Alert>Заявка сохранена и отправлена на модерацию.</Alert>
            </div>
          ) : null}

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
              <h2 className="text-xl font-semibold text-[var(--foreground)]">ПВЗ пока нет</h2>
              <p className="mx-auto mt-2 max-w-xl text-sm leading-6 text-[var(--muted)]">
                Добавьте первый ПВЗ, чтобы отправить его на модерацию.
              </p>
              <Link href="/owner/pvz/new" className="mt-5 inline-flex">
                <Button type="button">Добавить ПВЗ</Button>
              </Link>
            </div>
          ) : (
            <div className="mt-8 grid gap-4 lg:grid-cols-2">
              {items.map((pvz) => (
                <PvzCard key={pvz.id} pvz={pvz} />
              ))}
            </div>
          )}

          {nextPageToken ? (
            <div className="mt-6 flex justify-center">
              <Button
                type="button"
                variant="secondary"
                disabled={loadingMore}
                onClick={() => void load(nextPageToken)}
              >
                {loadingMore ? "Загружаем..." : "Показать еще"}
              </Button>
            </div>
          ) : null}
        </section>
      </div>
    </InventoryRoleGuard>
  );
}
