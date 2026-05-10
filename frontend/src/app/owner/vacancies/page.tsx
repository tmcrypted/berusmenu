import type { Metadata } from "next";
import Link from "next/link";
import { VacancyFilters } from "@/components/matching/VacancyFilters";
import { VacancyRow } from "@/components/matching/VacancyRow";
import { Button } from "@/components/ui/Button";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { listVacancies } from "@/lib/server-matching";
import { vacancyFiltersFromSearch } from "@/lib/matchingSearch";

export const metadata: Metadata = {
  title: "Мои вакансии | Беру Смену",
};

export default async function Page({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const auth = await getServerAuthSnapshot();
  const filters = vacancyFiltersFromSearch(await searchParams);
  const result = auth.user?.role === "owner" ? await listVacancies(filters) : null;

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Owner matching</p>
            <h1 className="mt-2 text-3xl font-semibold tracking-tight">Мои вакансии</h1>
          </div>
          <Link href="/owner/vacancies/new"><Button type="button">Создать вакансию</Button></Link>
        </div>
        {auth.user?.role !== "owner" ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Раздел доступен владельцу PVZ.</p>
        ) : (
          <>
            <div className="mt-6"><VacancyFilters filters={filters} /></div>
            {!result?.ok ? (
              <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{result?.message ?? "Не удалось загрузить вакансии."}</p>
            ) : result.data.items.length === 0 ? (
              <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-8 text-center text-sm text-[var(--muted)]">Вакансий пока нет.</p>
            ) : (
              <div className="mt-6 space-y-4">
                {result.data.items.map((vacancy) => (
                  <VacancyRow key={vacancy.id} vacancy={vacancy} user={auth.user} hrefBase="/owner/vacancies" />
                ))}
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
}
