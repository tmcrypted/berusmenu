import type { Metadata } from "next";
import Link from "next/link";
import { VacancyFilters } from "@/components/matching/VacancyFilters";
import { VacancyRow } from "@/components/matching/VacancyRow";
import { Button } from "@/components/ui/Button";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { listVacancies } from "@/lib/server-matching";
import { vacancyFiltersFromSearch } from "@/lib/matchingSearch";

export const metadata: Metadata = {
  title: "Вакансии | Беру Смену",
  description: "Лента срочных смен PVZ",
};

export default async function Page({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const auth = await getServerAuthSnapshot();
  const params = await searchParams;
  const filters = vacancyFiltersFromSearch(params);

  if (auth.status !== "authenticated_ready") {
    return (
      <div className="app-shell">
        <section className="mx-auto max-w-4xl px-4 py-10 sm:px-6 lg:px-8">
          <div className="rounded-[2rem] border border-[var(--border)] bg-[var(--panel)] p-8 shadow-[var(--shadow-soft)]">
            <h1 className="text-3xl font-semibold">Лента смен</h1>
            <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
              Войдите и завершите профиль, чтобы видеть реальные вакансии matching-service и откликаться.
            </p>
            <Link href="/login" className="mt-6 inline-flex">
              <Button type="button">Войти</Button>
            </Link>
          </div>
        </section>
      </div>
    );
  }

  const result = await listVacancies(filters);

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="mb-6">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Matching</p>
          <h1 className="mt-2 text-3xl font-semibold tracking-tight">Лента смен</h1>
          <p className="mt-2 text-sm text-[var(--muted)]">Горизонтальная лента вакансий с SSR-загрузкой и query-фильтрами.</p>
        </div>
        <VacancyFilters filters={filters} showStatus={auth.user?.role !== "worker"} showPvz={auth.user?.role !== "worker"} />
        {!result.ok ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{result.message}</p>
        ) : result.data.items.length === 0 ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center text-sm text-[var(--muted)]">Подходящих вакансий пока нет.</p>
        ) : (
          <div className="mt-6 space-y-4">
            {result.data.items.map((vacancy) => (
              <VacancyRow key={vacancy.id} vacancy={vacancy} user={auth.user} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
