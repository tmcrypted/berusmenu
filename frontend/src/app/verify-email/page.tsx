import type { Metadata } from "next";
import { VerifyEmailForm } from "@/components/auth/VerifyEmailForm";

export const metadata: Metadata = {
  title: "Подтверждение почты | Berusmenu",
  description: "Подтверждение email для доступа к рабочим сценариям Berusmenu",
};

export default function VerifyEmailPage() {
  return <VerifyEmailForm />;
}
