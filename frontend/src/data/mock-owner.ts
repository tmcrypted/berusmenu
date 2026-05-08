export type MockPvzCard = {
  id: string;
  title: string;
  address: string;
  statusLabel: string;
};

export type MockVacancyCard = {
  id: string;
  title: string;
  pvzLabel: string;
  windowLabel: string;
  payLabel: string;
};

/** Демо ПВЗ владельца (inventory-service позже). */
export const mockOwnerPvz: MockPvzCard[] = [
  {
    id: "demo-pvz-1",
    title: "ПВЗ Ozon · Тимирязевская",
    address: "Москва, ул. Яблочкова, 21к3",
    statusLabel: "На модерации",
  },
  {
    id: "demo-pvz-2",
    title: "ПВЗ Wildberries · Невский",
    address: "Санкт-Петербург, Невский пр., 100",
    statusLabel: "Одобрен",
  },
];

/** Демо вакансии этого владельца (matching-service позже). */
export const mockOwnerVacancies: MockVacancyCard[] = [
  {
    id: "demo-v-1",
    title: "Срочная подмена оперта",
    pvzLabel: "Ozon Тимирязевская",
    windowLabel: "Сегодня 10:00–22:00",
    payLabel: "2 500 ₽ / смена",
  },
  {
    id: "demo-v-2",
    title: "Старший смены",
    pvzLabel: "WB Невский",
    windowLabel: "Завтра 09:00–21:00",
    payLabel: "3 000 ₽",
  },
];
