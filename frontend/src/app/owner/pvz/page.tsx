import type { Metadata } from "next";
import { OwnerPvzPage } from "@/components/inventory/OwnerPvzPage";

export const metadata: Metadata = {
  title: "Мои ПВЗ | Беру Смену",
  description: "ПВЗ владельца и статусы модерации",
};

export default function Page() {
  return <OwnerPvzPage />;
}
