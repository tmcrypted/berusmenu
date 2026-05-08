import type { Metadata } from "next";
import { RegisterRoleChoice } from "@/components/auth/RegisterRoleChoice";

export const metadata: Metadata = {
  title: "Регистрация | Беру Смену . ру",
  description: "Выберите роль и создайте аккаунт в Berusmenu",
};

export default function RegisterPage() {
  return <RegisterRoleChoice />;
}
