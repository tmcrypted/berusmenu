import type { Metadata } from "next";
import { ModeratorPvzQueue } from "@/components/inventory/ModeratorPvzQueue";

export const metadata: Metadata = {
  title: "Модерация ПВЗ | Беру Смену",
  description: "Панель модератора для обработки заявок ПВЗ",
};

export default function Page() {
  return <ModeratorPvzQueue />;
}
