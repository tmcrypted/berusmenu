import type { Metadata } from "next";
import { RegisterForm } from "@/components/auth/RegisterForm";

export const metadata: Metadata = {
  title: "Регистрация | Беру Смену",
  description: "Создание аккаунта на Беру Смену",
};

export default function RegisterPage() {
  return <RegisterForm />;
}
