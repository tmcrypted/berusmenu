import type { AssignmentStatus, ResponseStatus, Vacancy, VacancyStatus } from "@/types/matching";

export function money(amount: number) {
  return new Intl.NumberFormat("ru-RU", { style: "currency", currency: "RUB", maximumFractionDigits: 0 }).format(amount);
}

export function dateTime(value: string) {
  if (!value) {
    return "Дата не указана";
  }
  return new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "short",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function dateOnly(value: string) {
  if (!value) {
    return "Дата не указана";
  }
  return new Intl.DateTimeFormat("ru-RU", {
    day: "numeric",
    month: "short",
  }).format(new Date(value));
}

export function shiftWindow(start: string, end: string) {
  return `${dateTime(start)} - ${dateTime(end)}`;
}

export function vacancyLocation(vacancy: Pick<Vacancy, "pvz_city_snapshot" | "pvz_address_snapshot">) {
  const city = vacancy.pvz_city_snapshot?.trim();
  const address = vacancy.pvz_address_snapshot.trim();

  if (!city) {
    return address;
  }
  if (!address || address.toLowerCase().includes(city.toLowerCase())) {
    return address || city;
  }
  return `${city}, ${address}`;
}

export function vacancyStatusLabel(status: VacancyStatus) {
  const labels: Record<VacancyStatus, string> = {
    published: "Опубликована",
    assigned: "Назначена",
    completed: "Завершена",
    cancelled: "Отменена",
    unspecified: "Неизвестно",
  };
  return labels[status];
}

export function responseStatusLabel(status: ResponseStatus) {
  const labels: Record<ResponseStatus, string> = {
    pending: "Ожидает",
    accepted: "Принят",
    rejected: "Отклонен",
    withdrawn: "Отозван",
    unspecified: "Неизвестно",
  };
  return labels[status];
}

export function assignmentStatusLabel(status: AssignmentStatus) {
  const labels: Record<AssignmentStatus, string> = {
    active: "Активно",
    completed: "Завершено",
    cancelled: "Отменено",
    unspecified: "Неизвестно",
  };
  return labels[status];
}
