import type { Metadata } from "next";
import { EditPvzPage } from "@/components/inventory/EditPvzPage";

export const metadata: Metadata = {
  title: "Редактировать ПВЗ | Беру Смену",
  description: "Редактирование ПВЗ и повторная отправка на модерацию",
};

export default async function Page({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return <EditPvzPage id={Number(id)} />;
}
