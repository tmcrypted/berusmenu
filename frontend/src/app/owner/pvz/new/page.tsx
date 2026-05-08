import type { Metadata } from "next";
import { InventoryRoleGuard } from "@/components/inventory/InventoryRoleGuard";
import { PvzForm } from "@/components/inventory/PvzForm";

export const metadata: Metadata = {
  title: "Новый ПВЗ | Беру Смену",
  description: "Подача ПВЗ на модерацию",
};

export default function NewPvzPage() {
  return (
    <InventoryRoleGuard role="owner">
      <div className="app-shell">
        <section className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8">
          <div className="rounded-2xl border border-[var(--border)] bg-[var(--card)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">
              Inventory
            </p>
            <h1 className="mt-2 text-2xl font-semibold tracking-tight text-[var(--foreground)]">
              Новый ПВЗ
            </h1>
            <p className="mt-2 text-sm leading-6 text-[var(--muted)]">
              Заполните данные точки. После отправки заявка попадет в очередь модератора.
            </p>
            <div className="mt-6">
              <PvzForm mode="create" />
            </div>
          </div>
        </section>
      </div>
    </InventoryRoleGuard>
  );
}
