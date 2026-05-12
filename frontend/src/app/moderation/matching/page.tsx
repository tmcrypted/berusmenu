import type { Metadata } from "next";
import { VacancyFilters } from "@/components/matching/VacancyFilters";
import { VacancyRow } from "@/components/matching/VacancyRow";
import { getServerAuthSnapshot } from "@/lib/server-auth";
import { listModerationVacancies } from "@/lib/server-matching";
import { vacancyFiltersFromSearch } from "@/lib/matchingSearch";

export const metadata: Metadata = {
  title: "Модерация matching | Беру Смену",
};

export default async function Page({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) {
  const auth = await getServerAuthSnapshot();
  const filters = vacancyFiltersFromSearch(await searchParams);
  const result = auth.user?.role === "moderator" ? await listModerationVacancies(filters) : null;

  return (
    <div className="app-shell">
      <section className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8">
        <p className="text-xs font-semibold uppercase tracking-[0.2em] text-[var(--accent)]">Moderation</p>
        <h1 className="mt-2 text-3xl font-semibold tracking-tight">Matching moderation</h1>
        {auth.user?.role !== "moderator" ? (
          <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--muted)]">Раздел доступен модератору.</p>
        ) : (
          <>
            <div className="mt-6"><VacancyFilters filters={filters} /></div>
            {!result?.ok ? (
              <p className="mt-6 rounded-2xl border border-[var(--border)] bg-[var(--card)] p-5 text-sm text-[var(--danger)]">{result?.message ?? "Не удалось загрузить вакансии."}</p>
            ) : (
              <div className="mt-6 space-y-4">
                {result.data.items.map((vacancy) => (
                  <VacancyRow key={vacancy.id} vacancy={vacancy} user={auth.user} />
                ))}
              </div>
            )}
          </>
        )}
      </section>
    </div>
  );
}
