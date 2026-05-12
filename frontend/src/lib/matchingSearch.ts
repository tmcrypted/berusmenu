import type { VacancyFilters, VacancyStatus } from "@/types/matching";

type Search = Record<string, string | string[] | undefined>;

export function vacancyFiltersFromSearch(searchParams: Search): VacancyFilters {
  const status = one(searchParams.status);
  const pvz = Number(one(searchParams.pvz_id));
  return {
    status: isVacancyStatus(status) ? status : undefined,
    pvz_id: Number.isFinite(pvz) && pvz > 0 ? pvz : undefined,
    shift_start_from: normalizeDate(one(searchParams.shift_start_from)),
    shift_start_to: normalizeDate(one(searchParams.shift_start_to)),
    page_token: one(searchParams.page_token) || undefined,
  };
}

function one(value: string | string[] | undefined) {
  return Array.isArray(value) ? value[0] : value;
}

function isVacancyStatus(value: string | undefined): value is VacancyStatus {
  return value === "published" || value === "assigned" || value === "completed" || value === "cancelled";
}

function normalizeDate(value: string | undefined) {
  if (!value) {
    return undefined;
  }
  if (value.includes("T") && !value.endsWith("Z") && !/[+-]\d\d:\d\d$/.test(value)) {
    return `${value}:00+03:00`;
  }
  return value;
}
