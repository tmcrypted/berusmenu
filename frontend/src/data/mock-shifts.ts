import type { ShiftCard } from "@/types/shift";

export const mockShifts: ShiftCard[] = [
  {
    id: "1",
    dateIso: "2026-04-05",
    dateLabel: "05.04.2026",
    role: "Работник ПВЗ",
    address: "г. Москва, ул. Примерная, д. 12, ТЦ «Север»",
    metro: "м. ВДНХ",
    timeWindow: "10:00 — 22:00",
    payLabel: "от 2 500 ₽ за смену",
    pvzName: "ПВЗ Ozon",
  },
  {
    id: "2",
    dateIso: "2026-04-05",
    dateLabel: "05.04.2026",
    role: "Работник выдачи заказов",
    address: "г. Санкт-Петербург, пр. Невский, д. 100",
    metro: "м. Невский проспект",
    timeWindow: "09:00 — 21:00",
    payLabel: "2 800 ₽",
    pvzName: "Wildberries",
  },
  {
    id: "3",
    dateIso: "2026-04-06",
    dateLabel: "06.04.2026",
    role: "Подмена администратора ПВЗ",
    address: "г. Казань, ул. Баумана, д. 58",
    metro: null,
    timeWindow: "11:00 — 20:00",
    payLabel: "договорная",
    pvzName: "СДЭК",
  },
];
