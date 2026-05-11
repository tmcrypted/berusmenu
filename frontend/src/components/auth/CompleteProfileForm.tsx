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
  last_name: z.string().trim().min(2, "Введите фамилию"),
  first_name: z.string().trim().min(2, "Введите имя"),
  middle_name: z.string().optional().transform((s) => (s ?? "").trim()),
  birth_date: z.string().trim().refine((s) => {
    const date = new Date(`${s}T00:00:00`);
    return s && !Number.isNaN(date.getTime()) && date < new Date();
  }, "Укажите дату рождения"),
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
  telegram_username: z.string().optional().transform((s) => (s ?? "").trim().replace(/^@+/, "")),
  max_username: z.string().optional().transform((s) => (s ?? "").trim().replace(/^@+/, "")),
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
      last_name: "",
      first_name: "",
      middle_name: "",
      birth_date: "",
      phone: "",
      avatar_url: "",
      telegram_username: "",
      max_username: "",
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
      last_name: user.last_name ?? "",
      first_name: user.first_name ?? "",
      middle_name: user.middle_name ?? "",
      birth_date: user.birth_date ?? "",
      phone: phoneDisplay,
      avatar_url: user.avatar_url ?? "",
      telegram_username: user.telegram_username ?? "",
      max_username: user.max_username ?? "",
    });
  }, [user, reset]);

  async function onSubmit(values: FormValues) {
    try {
      const updated = await updateMe({
        last_name: values.last_name,
        first_name: values.first_name,
        middle_name: values.middle_name || undefined,
        birth_date: values.birth_date,
        phone: values.phone,
        avatar_url: values.avatar_url?.trim() ? values.avatar_url : undefined,
        telegram_username: values.telegram_username || undefined,
        max_username: values.max_username || undefined,
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
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="Фамилия" htmlFor="complete-profile-last-name" error={errors.last_name?.message}>
                <Input id="complete-profile-last-name" autoComplete="family-name" placeholder="Иванов" {...register("last_name")} />
              </Field>
              <Field label="Имя" htmlFor="complete-profile-first-name" error={errors.first_name?.message}>
                <Input id="complete-profile-first-name" autoComplete="given-name" placeholder="Иван" {...register("first_name")} />
              </Field>
            </div>
            <Field label="Отчество" htmlFor="complete-profile-middle-name" hint="Необязательно" error={errors.middle_name?.message}>
              <Input id="complete-profile-middle-name" autoComplete="additional-name" placeholder="Иванович" {...register("middle_name")} />
            </Field>
            <Field label="Дата рождения" htmlFor="complete-profile-birth-date" error={errors.birth_date?.message}>
              <Input id="complete-profile-birth-date" type="date" {...register("birth_date")} />
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
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="Telegram" htmlFor="complete-profile-telegram" hint="Необязательно" error={errors.telegram_username?.message}>
                <Input id="complete-profile-telegram" placeholder="@username" {...register("telegram_username")} />
              </Field>
              <Field label="Max" htmlFor="complete-profile-max" hint="Необязательно" error={errors.max_username?.message}>
                <Input id="complete-profile-max" placeholder="username" {...register("max_username")} />
              </Field>
            </div>
            <Button type="submit" fullWidth disabled={isSubmitting}>
              {isSubmitting ? "Сохраняем..." : "Сохранить и продолжить"}
            </Button>
          </form>
        </div>
      </AuthShell>
    </AuthGuard>
  );
}
