"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { InventoryRoleGuard } from "@/components/inventory/InventoryRoleGuard";
import { PvzForm } from "@/components/inventory/PvzForm";
import { Alert } from "@/components/ui/Alert";
import { Button } from "@/components/ui/Button";
import { Spinner } from "@/components/ui/Spinner";
import { ApiError } from "@/lib/api";
import { getPvz } from "@/lib/inventory";
import type { Pvz } from "@/types/inventory";

export function EditPvzPage({ id }: { id: number }) {
  const router = useRouter();
  const [pvz, setPvz] = useState<Pvz | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function load() {
      try {
        setPvz(await getPvz(id));
      } catch (err) {
        setError(err instanceof ApiError ? err.message : "Не удалось загрузить ПВЗ.");
      } finally {
        setLoading(false);
      }
    }
    void load();
  }, [id]);

  return (
    <InventoryRoleGuard role="owner">
      <div className="app-shell">
        <section className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
          <Button type="button" variant="ghost" className="mb-5 rounded-xl" onClick={() => router.push("/owner/pvz")}>
            Назад к списку
          </Button>
          <div className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
            <h1 className="text-2xl font-semibold tracking-tight text-[var(--foreground)]">
              Редактировать ПВЗ
            </h1>
            <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
              После сохранения ПВЗ снова перейдет в статус модерации.
            </p>
            <div className="mt-6">
              {loading ? (
                <div className="flex justify-center py-12">
                  <Spinner />
                </div>
              ) : error ? (
                <Alert tone="error">{error}</Alert>
              ) : pvz ? (
                <PvzForm mode="edit" pvz={pvz} />
              ) : null}
            </div>
          </div>
        </section>
      </div>
    </InventoryRoleGuard>
  );
}
