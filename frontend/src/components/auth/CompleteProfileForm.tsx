"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { updateMe } from "@/lib/auth";
import { normalizeApiUser } from "@/lib/api-user";
import { ApiError } from "@/lib/api";
import { formatRuMobileDisplay, normalizeRussianMobile, RU_MOBILE_E164 } from "@/lib/phone-ru";
import { roleLabel } from "@/lib/profile";
import { useAuthStore } from "@/store/auth-store";
import type { UserRole } from "@/types/auth";
import { Alert } from "../ui/Alert";
import { Button } from "../ui/Button";
import { Field } from "../ui/Field";
import { Input } from "../ui/Input";
import { AuthGuard } from "./AuthGuard";
import { AuthShell } from "./AuthShell";

const schema = z.object({
  full_name: z.string().trim().min(2, "Введите имя или ФИО"),
  phone: z
    .string()
    .trim()
    .transform((s) => normalizeRussianMobile(s))
    .refine((s) => RU_MOBILE_E164.test(s), {
      message: "Телефон: мобильный РФ +7 9XX XXX-XX-XX (ровно 10 цифр после +7)",
    }),
  avatar_url: z
    .string()
    .optional()
    .transform((s) => (s ?? "").trim())
    .refine((s) => s === "" || /^https?:\/\/.+/i.test(s), {
      message: "Введите корректный URL или оставьте поле пустым",
    }),
});

type FormInput = z.input<typeof schema>;
type FormValues = z.output<typeof schema>;

type Props = {
  role: UserRole;
};

export function CompleteProfileForm({ role }: Props) {
  const router = useRouter();
  const user = useAuthStore((state) => state.user);
  const setReady = useAuthStore((state) => state.setReady);
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormInput, unknown, FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      full_name: "",
      phone: "",
      avatar_url: "",
    },
  });

  useEffect(() => {
    if (!user) {
      return;
    }
    const rawPhone = user.phone?.trim() ?? "";
    const normalized = rawPhone ? normalizeRussianMobile(rawPhone) : "";
    const phoneDisplay =
      normalized && RU_MOBILE_E164.test(normalized)
        ? formatRuMobileDisplay(normalized)
        : rawPhone;
    reset({
      full_name: user.full_name ?? "",
      phone: phoneDisplay,
      avatar_url: user.avatar_url ?? "",
    });
  }, [user, reset]);

  async function onSubmit(values: FormValues) {
    try {
      const updated = await updateMe({
        full_name: values.full_name,
        phone: values.phone,
        avatar_url: values.avatar_url?.trim() ? values.avatar_url : undefined,
      });

      setReady(normalizeApiUser(updated));
      router.push("/");
    } catch (error) {
      if (error instanceof ApiError) {
        setError("root", { message: error.message });
        return;
      }
      setError("root", {
        message: "Не удалось сохранить профиль. Попробуйте еще раз.",
      });
    }
  }

  const isOwner = role === "owner";

  return (
    <AuthGuard mode="profile_completion" role={role}>
      <AuthShell
        title={isOwner ? "Завершите профиль владельца" : "Завершите профиль исполнителя"}
        subtitle={
          isOwner
            ? "После этого можно переходить к owner-сценарию и подаче заявки на ПВЗ."
            : "После этого можно откликаться на смены и пользоваться полным доступом."
        }
      >
        <div className="space-y-5">
          <Alert>
            <p className="font-medium">Роль</p>
            <p className="mt-1 text-sm text-[var(--muted)]">{roleLabel(role)}</p>
          </Alert>
          {isOwner ? (
            <Alert>
              Следующий шаг после сохранения профиля: отправить заявку на одобрение ПВЗ.
              Полный owner-flow по ПВЗ остается следующим продуктовым этапом.
            </Alert>
          ) : null}
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
            {errors.root?.message ? <Alert tone="error">{errors.root.message}</Alert> : null}
            <Field
              label="Имя или ФИО"
              htmlFor="complete-profile-name"
              error={errors.full_name?.message}
            >
              <Input
                id="complete-profile-name"
                autoComplete="name"
                placeholder="Как к вам обращаться"
                {...register("full_name")}
              />
            </Field>
            <Field
              label="Телефон"
              htmlFor="complete-profile-phone"
              hint="Российский мобильный: +7 9XX XXX-XX-XX"
              error={errors.phone?.message}
            >
              <Input
                id="complete-profile-phone"
                type="tel"
                autoComplete="tel"
                placeholder="+7 (999) 123-45-67"
                {...register("phone")}
              />
            </Field>
            <Field
              label="Ссылка на аватар"
              htmlFor="complete-profile-avatar"
              hint="Необязательно. Можно добавить позже."
              error={errors.avatar_url?.message}
            >
              <Input
                id="complete-profile-avatar"
                type="url"
                autoComplete="url"
                placeholder="https://example.com/avatar.jpg"
                {...register("avatar_url")}
              />
            </Field>
            <Button type="submit" fullWidth disabled={isSubmitting}>
              {isSubmitting ? "Сохраняем..." : "Сохранить и продолжить"}
            </Button>
          </form>
        </div>
      </AuthShell>
    </AuthGuard>
  );
}
