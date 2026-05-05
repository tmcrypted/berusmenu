import type { Metadata } from "next";
import { LoginForm } from "@/components/auth/LoginForm";

export const metadata: Metadata = {
  title: "Вход | Беру Смену",
  description: "Вход в личный кабинет Беру Смену",
};

export default function LoginPage() {
  return <LoginForm />;
}
