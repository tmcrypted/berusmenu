"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { register as registerUser } from "@/lib/auth";
import { ApiError } from "@/lib/api";
import { resolveApiAuthSnapshot } from "@/lib/auth-state";
import { useAuthStore } from "@/store/auth-store";
import { Alert } from "../ui/Alert";
import { Button } from "../ui/Button";
import { Field } from "../ui/Field";
import { Input } from "../ui/Input";
import { PasswordInput } from "../ui/PasswordInput";
import { AuthGuard } from "./AuthGuard";
import { AuthShell } from "./AuthShell";

const schema = z
  .object({
    email: z.email("Введите корректный email"),
    password: z.string().min(8, "Пароль должен быть не короче 8 символов"),
    confirmPassword: z.string().min(8, "Повторите пароль"),
    role: z.enum(["worker", "owner"]),
  })
  .refine((values) => values.password === values.confirmPassword, {
    path: ["confirmPassword"],
    message: "Пароли не совпадают",
  });

type FormValues = z.infer<typeof schema>;

type Props = {
  initialRole: "worker" | "owner";
};

export function RegisterForm({ initialRole }: Props) {
  const router = useRouter();
  const hydrate = useAuthStore((state) => state.hydrate);
  const roleSubtitle =
    initialRole === "owner"
      ? "Вы регистрируетесь как владелец ПВЗ. Дальше — почта, пароль и подтверждение email."
      : "Вы регистрируетесь как исполнитель. Дальше — почта, пароль и подтверждение email.";
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
    setError,
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      email: "",
      password: "",
      confirmPassword: "",
      role: initialRole,
    },
  });

  async function onSubmit(values: FormValues) {
    try {
      const auth = await registerUser({
        email: values.email,
        password: values.password,
        role: values.role,
      });

      hydrate(resolveApiAuthSnapshot(auth.user, auth.access_expires_at));
      router.push("/verify-email");
    } catch (error) {
      if (error instanceof ApiError) {
        setError("root", { message: error.message });
        return;
      }
      setError("root", {
        message: "Не удалось создать аккаунт. Попробуйте еще раз.",
      });
    }
  }

  return (
    <AuthGuard mode="anonymous">
      <AuthShell
        title="Регистрация"
        subtitle={roleSubtitle}
        footer={
          <>
            <div className="mb-4">
              <Link
                href="/register"
                className="text-sm font-medium text-[var(--accent)] hover:text-[var(--accent-hover)]"
              >
                ← Изменить выбор роли
              </Link>
            </div>
            Уже есть аккаунт?{" "}
            <Link
              href="/login"
              className="font-semibold text-[var(--accent)] hover:text-[var(--accent-hover)]"
            >
              Войти
            </Link>
          </>
        }
      >
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          {errors.root?.message ? <Alert tone="error">{errors.root.message}</Alert> : null}
          <Field label="Email" htmlFor="register-email" error={errors.email?.message}>
            <Input
              id="register-email"
              type="email"
              autoComplete="email"
              placeholder="you@example.com"
              {...register("email")}
            />
          </Field>
          <input type="hidden" {...register("role")} />
          <Field label="Пароль" htmlFor="register-password" error={errors.password?.message}>
            <PasswordInput
              id="register-password"
              autoComplete="new-password"
              placeholder="Минимум 8 символов"
              {...register("password")}
            />
          </Field>
          <Field
            label="Повторите пароль"
            htmlFor="register-confirm-password"
            error={errors.confirmPassword?.message}
          >
            <PasswordInput
              id="register-confirm-password"
              autoComplete="new-password"
              placeholder="Повторите пароль"
              {...register("confirmPassword")}
            />
          </Field>
          <Button type="submit" fullWidth disabled={isSubmitting}>
            {isSubmitting ? "Создаем аккаунт..." : "Продолжить"}
          </Button>
        </form>
      </AuthShell>
    </AuthGuard>
  );
}
