import type { Metadata } from "next";
import { CompleteProfileForm } from "@/components/auth/CompleteProfileForm";

export const metadata: Metadata = {
  title: "Профиль исполнителя | Berusmenu",
  description: "Заполнение профиля исполнителя перед откликами на смены",
};

export default function CompleteWorkerProfilePage() {
  return <CompleteProfileForm role="worker" />;
}
