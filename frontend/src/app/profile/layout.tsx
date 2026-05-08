import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Профиль | Беру Смену . ру",
  description: "Данные аккаунта и роль в Berusmenu",
};

export default function ProfileLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
