import type { Metadata } from "next";
import { VacancyForm } from "@/components/matching/VacancyForm";

export const metadata: Metadata = {
  title: "Новая вакансия | Беру Смену",
};

export default function Page() {
  return (
    <div className="app-shell">
      <section className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-6 shadow-[var(--shadow-soft)] sm:p-8">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Owner matching</p>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">Создать вакансию</h1>
          <p className="mt-2 text-sm text-[var(--muted)]">Вакансию можно создать только для approved PVZ.</p>
          <div className="mt-6">
            <VacancyForm />
          </div>
        </div>
      </section>
    </div>
  );
}
