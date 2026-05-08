import type { Metadata } from "next";
import { LoginForm } from "@/components/auth/LoginForm";

export const metadata: Metadata = {
  title: "Вход | Berusmenu",
  description: "Вход в Berusmenu для исполнителей и владельцев ПВЗ",
};

export default function LoginPage() {
  return <LoginForm />;
}
