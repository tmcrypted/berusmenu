import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { RegisterForm } from "@/components/auth/RegisterForm";

const ROLES = ["worker", "owner"] as const;
type Role = (typeof ROLES)[number];

export async function generateStaticParams() {
  return ROLES.map((role) => ({ role }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ role: string }>;
}): Promise<Metadata> {
  const { role } = await params;
  if (!isRole(role)) {
    return { title: "Регистрация | Беру Смену . ру" };
  }
  const label = role === "owner" ? "владелец ПВЗ" : "исполнитель";
  return {
    title: `Регистрация (${label}) | Беру Смену . ру`,
    description: "Создание аккаунта в Berusmenu: email и пароль",
  };
}

function isRole(s: string): s is Role {
  return ROLES.includes(s as Role);
}

export default async function RegisterWithRolePage({
  params,
}: {
  params: Promise<{ role: string }>;
}) {
  const { role } = await params;
  if (!isRole(role)) {
    notFound();
  }
  return <RegisterForm initialRole={role} />;
}
