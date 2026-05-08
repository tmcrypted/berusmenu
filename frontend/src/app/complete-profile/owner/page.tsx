import type { Metadata } from "next";
import { CompleteProfileForm } from "@/components/auth/CompleteProfileForm";

export const metadata: Metadata = {
  title: "Профиль владельца | Berusmenu",
  description: "Заполнение профиля владельца перед PVZ-потоком",
};

export default function CompleteOwnerProfilePage() {
  return <CompleteProfileForm role="owner" />;
}
